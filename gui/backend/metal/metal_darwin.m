#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>
#include "metal_darwin.h"
#include <string.h>

// ─── Global State ─────────────────────────────────────────────

static id<MTLDevice>       _device;
static id<MTLCommandQueue> _queue;
static CAMetalLayer        *_layer;
static id<MTLSamplerState> _sampler;
static id<MTLBuffer>       _quadIdx;

static id<MTLRenderPipelineState> _pipelines[PIPE_COUNT];

// Per-frame state.
static id<CAMetalDrawable>          _drawable;
static id<MTLCommandBuffer>         _cmdBuf;
static id<MTLRenderCommandEncoder>  _enc;

// Viewport.
static int _viewW, _viewH;

// Textures.
#define MAX_TEX 8192
static id<MTLTexture> _textures[MAX_TEX];
static int _nextTexID = 1;
static int _freeTexIDs[MAX_TEX];
static int _freeTexCount = 0;

// Reusable large-triangle upload buffers (triple-buffered per frame
// to avoid per-draw allocations and CPU/GPU write hazards).
#define TRI_BUF_RING 3
#define TRI_BUF_MAX_PER_FRAME 256
static id<MTLBuffer> _triBufs[TRI_BUF_RING][TRI_BUF_MAX_PER_FRAME];
static int _triBufCursor[TRI_BUF_RING];
static int _triBufFrame = -1;

// Filter textures.
static id<MTLTexture> _filterTexA;
static id<MTLTexture> _filterTexB;
static int _filterW, _filterH;

// ─── MSL Shader Source ────────────────────────────────────────

static NSString *mslSource = @R"(
#include <metal_stdlib>
using namespace metal;

// ── Common Structs ──

struct VertexIn {
    float3 position [[attribute(0)]];
    float2 texcoord [[attribute(1)]];
    float4 color    [[attribute(2)]];
};

struct VertexOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
    float  params;
};

struct ShadowOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
    float  params;
    float2 offset;
};

struct GradientOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
    float  params;
    float4 stop12;
    float4 stop34;
    float4 stop56;
    float4 meta;
};

struct FilterOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
    float  std_dev;
};

struct GlyphIn {
    float2 position [[attribute(0)]];
    float2 texcoord [[attribute(1)]];
    float4 color    [[attribute(2)]];
};

struct GlyphOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
};

// ── Vertex Shaders ──

vertex VertexOut vs_solid(
    VertexIn in [[stage_in]],
    constant float4x4 &mvp [[buffer(1)]]
) {
    VertexOut out;
    out.position = mvp * float4(in.position.xy, 0.0, 1.0);
    out.uv       = in.texcoord;
    out.color    = in.color;
    out.params   = in.position.z;
    return out;
}

vertex ShadowOut vs_shadow(
    VertexIn in [[stage_in]],
    constant float4x4 &mvp [[buffer(1)]],
    constant float4x4 &tm  [[buffer(2)]]
) {
    ShadowOut out;
    out.position = mvp * float4(in.position.xy, 0.0, 1.0);
    out.uv       = in.texcoord;
    out.color    = in.color;
    out.params   = in.position.z;
    out.offset   = (tm * float4(0, 0, 0, 1)).xy;
    return out;
}

vertex GradientOut vs_gradient(
    VertexIn in [[stage_in]],
    constant float4x4 &mvp [[buffer(1)]],
    constant float4x4 &tm  [[buffer(2)]]
) {
    GradientOut out;
    out.position = mvp * float4(in.position.xy, 0.0, 1.0);
    out.uv       = in.texcoord;
    out.color    = in.color;
    out.params   = in.position.z;
    out.stop12   = tm[0];
    out.stop34   = tm[1];
    out.stop56   = tm[2];
    out.meta     = tm[3];
    return out;
}

vertex FilterOut vs_filter(
    VertexIn in [[stage_in]],
    constant float4x4 &mvp [[buffer(1)]],
    constant float4x4 &tm  [[buffer(2)]]
) {
    FilterOut out;
    out.position = mvp * float4(in.position.xy, 0.0, 1.0);
    out.uv       = in.texcoord;
    out.color    = in.color;
    out.std_dev  = tm[0][0];
    return out;
}

vertex GlyphOut vs_glyph(
    GlyphIn in [[stage_in]],
    constant float4x4 &mvp [[buffer(1)]]
) {
    GlyphOut out;
    out.position = mvp * float4(in.position, 0.0, 1.0);
    out.uv       = in.texcoord;
    out.color    = in.color;
    return out;
}

// ── Fragment Shaders ──

fragment float4 fs_solid(VertexOut in [[stage_in]]) {
    float radius    = floor(in.params / 4096.0) / 4.0;
    float thickness = fmod(in.params, 4096.0) / 4.0;

    float2 uv_to_px  = 1.0 / (float2(fwidth(in.uv.x),
                        fwidth(in.uv.y)) + 1e-6);
    float2 half_size  = uv_to_px;
    float2 pos        = in.uv * half_size;

    float2 q = abs(pos) - half_size + float2(radius);
    float d = length(max(q, float2(0.0)))
            + min(max(q.x, q.y), 0.0) - radius;

    if (thickness > 0.0) {
        d = abs(d + thickness * 0.5) - thickness * 0.5;
    }

    float grad_len = length(float2(dfdx(d), dfdy(d)));
    d = d / max(grad_len, 0.001);
    float alpha = 1.0 - smoothstep(-0.59, 0.59, d);
    return float4(in.color.rgb, in.color.a * alpha);
}

fragment float4 fs_shadow(ShadowOut in [[stage_in]]) {
    float radius = floor(in.params / 4096.0) / 4.0;
    float blur   = fmod(in.params, 4096.0) / 4.0;

    float2 uv_to_px = 1.0 / (float2(fwidth(in.uv.x),
                       fwidth(in.uv.y)) + 1e-6);
    float2 half_size = uv_to_px;
    float2 pos       = in.uv * half_size;

    float2 q = abs(pos) - half_size
             + float2(radius + 1.5 * blur);
    float d = length(max(q, float2(0.0)))
            + min(max(q.x, q.y), 0.0) - radius;

    float2 q_c = abs(pos + in.offset) - half_size
               + float2(radius + 1.5 * blur);
    float d_c = length(max(q_c, float2(0.0)))
              + min(max(q_c.x, q_c.y), 0.0) - radius;

    float alpha_falloff = 1.0 - smoothstep(0.0,
                          max(1.0, blur), d);
    float alpha_clip    = smoothstep(-1.0, 0.0, d_c);
    float alpha         = alpha_falloff * alpha_clip;

    return float4(in.color.rgb, in.color.a * alpha);
}

fragment float4 fs_blur(ShadowOut in [[stage_in]]) {
    float radius = floor(in.params / 4096.0) / 4.0;
    float blur   = fmod(in.params, 4096.0) / 4.0;

    float2 uv_to_px = 1.0 / (float2(fwidth(in.uv.x),
                       fwidth(in.uv.y)) + 1e-6);
    float2 half_size = uv_to_px;
    float2 pos       = in.uv * half_size;

    float2 q = abs(pos) - half_size
             + float2(radius + 1.5 * blur);
    float d = length(max(q, float2(0.0)))
            + min(max(q.x, q.y), 0.0) - radius;

    float alpha = 1.0 - smoothstep(-blur, blur, d);
    return float4(in.color.rgb, in.color.a * alpha);
}

fragment float4 fs_gradient(GradientOut in [[stage_in]]) {
    float radius = floor(in.params / 4096.0) / 4.0;

    float hw = in.meta.x;
    float hh = in.meta.y;
    float grad_type = in.meta.z;
    int stop_count  = int(in.meta.w);

    float2 pos = in.uv * float2(hw, hh);

    float2 q = abs(pos) - float2(hw, hh) + float2(radius);
    float d = length(max(q, float2(0.0)))
            + min(max(q.x, q.y), 0.0) - radius;
    float sdf_alpha = 1.0 - smoothstep(-0.5, 0.5, d);

    float t;
    if (grad_type > 0.5) {
        float target_radius = in.stop56.w;
        t = length(pos) / target_radius;
    } else {
        float2 stop_dir = float2(in.stop56.z, in.stop56.w);
        t = dot(in.uv, stop_dir) * 0.5 + 0.5;
    }
    t = clamp(t, 0.0, 1.0);

    // Unpack gradient stops.
    float4 stop_colors[6];
    float  stop_positions[6];

    // Helper lambda-like macros via inline.
    auto unpack = [](float val1, float val2,
                     thread float4 &c, thread float &p) {
        float r = fmod(val1, 256.0);
        float g = fmod(floor(val1 / 256.0), 256.0);
        float b = floor(val1 / 65536.0);
        float a = fmod(val2, 256.0);
        p = floor(val2 / 256.0) / 10000.0;
        c = float4(r/255.0, g/255.0, b/255.0, a/255.0);
    };

    unpack(in.stop12.x, in.stop12.y,
           stop_colors[0], stop_positions[0]);
    unpack(in.stop12.z, in.stop12.w,
           stop_colors[1], stop_positions[1]);
    unpack(in.stop34.x, in.stop34.y,
           stop_colors[2], stop_positions[2]);
    unpack(in.stop34.z, in.stop34.w,
           stop_colors[3], stop_positions[3]);
    unpack(in.stop56.x, in.stop56.y,
           stop_colors[4], stop_positions[4]);
    stop_colors[5]    = stop_colors[4];
    stop_positions[5] = stop_positions[4];

    float4 c1 = stop_colors[0];
    float4 c2 = c1;
    float  p1 = stop_positions[0];
    float  p2 = p1;

    for (int i = 1; i < 6; i++) {
        if (i >= stop_count) break;
        if (t <= stop_positions[i]) {
            c2 = stop_colors[i];
            p2 = stop_positions[i];
            c1 = stop_colors[i-1];
            p1 = stop_positions[i-1];
            break;
        }
        if (i == stop_count - 1) {
            c1 = stop_colors[i];
            c2 = c1;
            p1 = stop_positions[i];
            p2 = p1;
        }
    }

    float local_t = (t - p1) / max(p2 - p1, 0.0001);

    float3 c1_pre  = c1.rgb * c1.a;
    float3 c2_pre  = c2.rgb * c2.a;
    float3 rgb_pre = mix(c1_pre, c2_pre, local_t);
    float  alpha   = mix(c1.a, c2.a, local_t);
    float3 rgb     = rgb_pre / max(alpha, 0.0001);
    float4 gc      = float4(rgb, alpha);

    // Dithering to reduce banding.
    float2 fc = float2(in.position.xy);
    float dither = fract(sin(dot(fc, float2(12.9898, 78.233)))
                   * 43758.5453) - 0.5;
    gc.rgb += float3(dither / 255.0);

    return float4(gc.rgb, gc.a * sdf_alpha * in.color.a);
}

fragment float4 fs_image_clip(
    VertexOut in [[stage_in]],
    texture2d<float> tex [[texture(0)]],
    sampler smp [[sampler(0)]]
) {
    float radius = floor(in.params / 4096.0) / 4.0;

    float2 uv_to_px = 1.0 / (float2(fwidth(in.uv.x),
                       fwidth(in.uv.y)) + 1e-6);
    float2 half_size = uv_to_px;
    float2 pos       = in.uv * half_size;

    float2 q = abs(pos) - half_size + float2(radius);
    float d = length(max(q, float2(0.0)))
            + min(max(q.x, q.y), 0.0) - radius;

    float grad_len = length(float2(dfdx(d), dfdy(d)));
    d = d / max(grad_len, 0.001);
    float alpha = 1.0 - smoothstep(-0.59, 0.59, d);

    float2 tex_uv   = in.uv * 0.5 + 0.5;
    float4 tex_color = tex.sample(smp, tex_uv);
    return float4(tex_color.rgb, tex_color.a * alpha);
}

fragment float4 fs_filter_blur_h(
    FilterOut in [[stage_in]],
    texture2d<float> tex [[texture(0)]],
    sampler smp [[sampler(0)]]
) {
    constexpr float w[7] = {
        0.19947, 0.17603, 0.12098,
        0.06476, 0.02700, 0.00877, 0.00222
    };
    float tex_w    = tex.get_width();
    float step_sz  = in.std_dev / tex_w;

    float4 c = tex.sample(smp, in.uv) * w[0];
    for (int i = 1; i < 7; i++) {
        float off = float(i) * step_sz;
        c += tex.sample(smp, in.uv + float2(off, 0)) * w[i];
        c += tex.sample(smp, in.uv - float2(off, 0)) * w[i];
    }
    return c;
}

fragment float4 fs_filter_blur_v(
    FilterOut in [[stage_in]],
    texture2d<float> tex [[texture(0)]],
    sampler smp [[sampler(0)]]
) {
    constexpr float w[7] = {
        0.19947, 0.17603, 0.12098,
        0.06476, 0.02700, 0.00877, 0.00222
    };
    float tex_h    = tex.get_height();
    float step_sz  = in.std_dev / tex_h;

    float4 c = tex.sample(smp, in.uv) * w[0];
    for (int i = 1; i < 7; i++) {
        float off = float(i) * step_sz;
        c += tex.sample(smp, in.uv + float2(0, off)) * w[i];
        c += tex.sample(smp, in.uv - float2(0, off)) * w[i];
    }
    return c;
}

fragment float4 fs_filter_tex(
    FilterOut in [[stage_in]],
    texture2d<float> tex [[texture(0)]],
    sampler smp [[sampler(0)]]
) {
    return tex.sample(smp, in.uv) * in.color;
}

fragment float4 fs_filter_color(FilterOut in [[stage_in]]) {
    return in.color;
}

fragment float4 fs_glyph_tex(
    GlyphOut in [[stage_in]],
    texture2d<float> tex [[texture(0)]],
    sampler smp [[sampler(0)]]
) {
    return tex.sample(smp, in.uv) * in.color;
}

fragment float4 fs_glyph_color(GlyphOut in [[stage_in]]) {
    return in.color;
}
)";

// ─── Helpers ──────────────────────────────────────────────────

static MTLVertexDescriptor *mainVertexDesc(void) {
    MTLVertexDescriptor *d = [[MTLVertexDescriptor alloc] init];
    // float3 position
    d.attributes[0].format      = MTLVertexFormatFloat3;
    d.attributes[0].offset      = 0;
    d.attributes[0].bufferIndex = 0;
    // float2 texcoord
    d.attributes[1].format      = MTLVertexFormatFloat2;
    d.attributes[1].offset      = 12;
    d.attributes[1].bufferIndex = 0;
    // float4 color
    d.attributes[2].format      = MTLVertexFormatFloat4;
    d.attributes[2].offset      = 20;
    d.attributes[2].bufferIndex = 0;
    d.layouts[0].stride         = 36;
    d.layouts[0].stepFunction   = MTLVertexStepFunctionPerVertex;
    return d;
}

static MTLVertexDescriptor *glyphVertexDesc(void) {
    MTLVertexDescriptor *d = [[MTLVertexDescriptor alloc] init];
    // float2 position
    d.attributes[0].format      = MTLVertexFormatFloat2;
    d.attributes[0].offset      = 0;
    d.attributes[0].bufferIndex = 0;
    // float2 texcoord
    d.attributes[1].format      = MTLVertexFormatFloat2;
    d.attributes[1].offset      = 8;
    d.attributes[1].bufferIndex = 0;
    // float4 color
    d.attributes[2].format      = MTLVertexFormatFloat4;
    d.attributes[2].offset      = 16;
    d.attributes[2].bufferIndex = 0;
    d.layouts[0].stride         = 32;
    d.layouts[0].stepFunction   = MTLVertexStepFunctionPerVertex;
    return d;
}

static id<MTLRenderPipelineState> makePipeline(
    id<MTLLibrary> lib,
    NSString *vsName,
    NSString *fsName,
    MTLVertexDescriptor *vd,
    MTLPixelFormat pixFmt
) {
    MTLRenderPipelineDescriptor *desc =
        [[MTLRenderPipelineDescriptor alloc] init];
    desc.vertexFunction   = [lib newFunctionWithName:vsName];
    desc.fragmentFunction = [lib newFunctionWithName:fsName];
    desc.vertexDescriptor = vd;

    desc.colorAttachments[0].pixelFormat = pixFmt;
    desc.colorAttachments[0].blendingEnabled = YES;
    desc.colorAttachments[0].sourceRGBBlendFactor =
        MTLBlendFactorSourceAlpha;
    desc.colorAttachments[0].destinationRGBBlendFactor =
        MTLBlendFactorOneMinusSourceAlpha;
    desc.colorAttachments[0].sourceAlphaBlendFactor =
        MTLBlendFactorSourceAlpha;
    desc.colorAttachments[0].destinationAlphaBlendFactor =
        MTLBlendFactorOneMinusSourceAlpha;

    if (!desc.vertexFunction) {
        NSLog(@"metal: vertex function %@ not found", vsName);
        return nil;
    }
    if (!desc.fragmentFunction) {
        NSLog(@"metal: fragment function %@ not found", fsName);
        return nil;
    }

    NSError *err = nil;
    id<MTLRenderPipelineState> pso =
        [_device newRenderPipelineStateWithDescriptor:desc
                                                error:&err];
    if (!pso) {
        NSLog(@"metal: pipeline %@/%@: %@", vsName, fsName, err);
    }
    return pso;
}

static id<MTLTexture> makeTexture(int w, int h,
    MTLPixelFormat fmt) {
    MTLTextureDescriptor *td =
        [MTLTextureDescriptor texture2DDescriptorWithPixelFormat:fmt
                              width:w height:h mipmapped:NO];
    td.usage = MTLTextureUsageShaderRead;
    td.storageMode = MTLStorageModeShared;
    return [_device newTextureWithDescriptor:td];
}

static id<MTLTexture> makeRenderTarget(int w, int h,
    MTLPixelFormat fmt) {
    MTLTextureDescriptor *td =
        [MTLTextureDescriptor texture2DDescriptorWithPixelFormat:fmt
                              width:w height:h mipmapped:NO];
    td.usage = MTLTextureUsageShaderRead |
               MTLTextureUsageRenderTarget;
    td.storageMode = MTLStorageModePrivate;
    return [_device newTextureWithDescriptor:td];
}

// Start or resume the main render pass.
static void beginMainEncoder(float r, float g, float b,
    float a, int clear) {
    MTLRenderPassDescriptor *rpd =
        [MTLRenderPassDescriptor renderPassDescriptor];
    rpd.colorAttachments[0].texture = _drawable.texture;
    rpd.colorAttachments[0].storeAction = MTLStoreActionStore;
    if (clear) {
        rpd.colorAttachments[0].loadAction = MTLLoadActionClear;
        rpd.colorAttachments[0].clearColor =
            MTLClearColorMake(r, g, b, a);
    } else {
        rpd.colorAttachments[0].loadAction = MTLLoadActionLoad;
    }
    _enc = [_cmdBuf renderCommandEncoderWithDescriptor:rpd];
    [_enc setViewport:(MTLViewport){
        0, 0, (double)_viewW, (double)_viewH, 0, 1}];
    [_enc setFragmentSamplerState:_sampler atIndex:0];
}

// ─── Public API ───────────────────────────────────────────────

int metalInit(void* layerPtr) {
    _layer = (__bridge CAMetalLayer*)layerPtr;
    _device = MTLCreateSystemDefaultDevice();
    if (!_device) {
        NSLog(@"metal: no Metal device");
        return -1;
    }

    _layer.device = _device;
    _layer.pixelFormat = MTLPixelFormatBGRA8Unorm;
    _layer.framebufferOnly = YES;
    // Synchronize presentation with the compositor resize
    // transaction. Eliminates content shift during live resize.
    _layer.presentsWithTransaction = YES;

    _queue = [_device newCommandQueue];

    // Compile MSL library.
    NSError *err = nil;
    id<MTLLibrary> lib =
        [_device newLibraryWithSource:mslSource
                              options:nil error:&err];
    if (!lib) {
        NSLog(@"metal: compile shaders: %@", err);
        return -2;
    }

    MTLPixelFormat pf = MTLPixelFormatBGRA8Unorm;
    MTLVertexDescriptor *mvd = mainVertexDesc();
    MTLVertexDescriptor *gvd = glyphVertexDesc();

    // Build pipeline states.
    _pipelines[PIPE_SOLID] =
        makePipeline(lib, @"vs_solid", @"fs_solid", mvd, pf);
    _pipelines[PIPE_SHADOW] =
        makePipeline(lib, @"vs_shadow", @"fs_shadow", mvd, pf);
    _pipelines[PIPE_BLUR] =
        makePipeline(lib, @"vs_shadow", @"fs_blur", mvd, pf);
    _pipelines[PIPE_GRADIENT] =
        makePipeline(lib, @"vs_gradient", @"fs_gradient",
                     mvd, pf);
    _pipelines[PIPE_IMAGE_CLIP] =
        makePipeline(lib, @"vs_solid", @"fs_image_clip",
                     mvd, pf);
    _pipelines[PIPE_FILTER_BLUR_H] =
        makePipeline(lib, @"vs_filter", @"fs_filter_blur_h",
                     mvd, pf);
    _pipelines[PIPE_FILTER_BLUR_V] =
        makePipeline(lib, @"vs_filter", @"fs_filter_blur_v",
                     mvd, pf);
    _pipelines[PIPE_FILTER_TEX] =
        makePipeline(lib, @"vs_filter", @"fs_filter_tex",
                     mvd, pf);
    _pipelines[PIPE_FILTER_COLOR] =
        makePipeline(lib, @"vs_filter", @"fs_filter_color",
                     mvd, pf);
    _pipelines[PIPE_GLYPH_TEX] =
        makePipeline(lib, @"vs_glyph", @"fs_glyph_tex",
                     gvd, pf);
    _pipelines[PIPE_GLYPH_COLOR] =
        makePipeline(lib, @"vs_glyph", @"fs_glyph_color",
                     gvd, pf);

    for (int i = 0; i < PIPE_COUNT; i++) {
        if (!_pipelines[i]) return -3;
    }

    // Quad index buffer: two triangles [0,1,2, 0,2,3].
    uint16_t idx[6] = {0, 1, 2, 0, 2, 3};
    _quadIdx = [_device newBufferWithBytes:idx
                                    length:sizeof(idx)
                                   options:MTLResourceStorageModeShared];

    // Shared sampler (linear + clamp-to-edge).
    MTLSamplerDescriptor *sd = [[MTLSamplerDescriptor alloc] init];
    sd.minFilter    = MTLSamplerMinMagFilterLinear;
    sd.magFilter    = MTLSamplerMinMagFilterLinear;
    sd.sAddressMode = MTLSamplerAddressModeClampToEdge;
    sd.tAddressMode = MTLSamplerAddressModeClampToEdge;
    _sampler = [_device newSamplerStateWithDescriptor:sd];

    return 0;
}

void metalDestroy(void) {
    for (int i = 0; i < MAX_TEX; i++) {
        _textures[i] = nil;
    }
    _nextTexID = 1;
    _freeTexCount = 0;
    for (int f = 0; f < TRI_BUF_RING; f++) {
        _triBufCursor[f] = 0;
        for (int i = 0; i < TRI_BUF_MAX_PER_FRAME; i++) {
            _triBufs[f][i] = nil;
        }
    }
    _triBufFrame = -1;
    _filterTexA = nil;
    _filterTexB = nil;
    for (int i = 0; i < PIPE_COUNT; i++) {
        _pipelines[i] = nil;
    }
    _quadIdx = nil;
    _sampler = nil;
    _queue   = nil;
    _device  = nil;
    _layer   = nil;
}

void metalResize(int w, int h) {
    _viewW = w;
    _viewH = h;
    _layer.drawableSize = CGSizeMake(w, h);
}

int metalBeginFrame(float r, float g, float b, float a) {
    @autoreleasepool {
        _drawable = [_layer nextDrawable];
    }
    if (!_drawable) return -1;

    _triBufFrame = (_triBufFrame + 1) % TRI_BUF_RING;
    _triBufCursor[_triBufFrame] = 0;

    _cmdBuf = [_queue commandBuffer];
    beginMainEncoder(r, g, b, a, 1);
    return 0;
}

void metalEndFrame(void) {
    if (_enc) {
        [_enc endEncoding];
        _enc = nil;
    }
    if (_drawable && _cmdBuf) {
        [_cmdBuf commit];
        [_cmdBuf waitUntilScheduled];
        [_drawable present];
    }
    _drawable = nil;
    _cmdBuf   = nil;
}

void metalSetPipeline(int id) {
    if (id < 0 || id >= PIPE_COUNT || !_enc) return;
    [_enc setRenderPipelineState:_pipelines[id]];
}

void metalSetMVP(const float* m) {
    if (!_enc) return;
    [_enc setVertexBytes:m length:64 atIndex:1];
}

void metalSetTM(const float* m) {
    if (!_enc) return;
    [_enc setVertexBytes:m length:64 atIndex:2];
}

void metalSetScissor(int x, int y, int w, int h, int viewH) {
    if (!_enc) return;
    // Clamp to viewport.
    if (x < 0) { w += x; x = 0; }
    if (y < 0) { h += y; y = 0; }
    if (w <= 0 || h <= 0) {
        // Zero-area scissor: clip everything.
        [_enc setScissorRect:(MTLScissorRect){0, 0, 1, 1}];
        return;
    }
    if (x + w > _viewW) w = _viewW - x;
    if (y + h > _viewH) h = _viewH - y;
    if (w <= 0 || h <= 0) {
        [_enc setScissorRect:(MTLScissorRect){0, 0, 1, 1}];
        return;
    }
    [_enc setScissorRect:(MTLScissorRect){
        (NSUInteger)x, (NSUInteger)y,
        (NSUInteger)w, (NSUInteger)h}];
}

void metalDisableScissor(void) {
    if (!_enc) return;
    [_enc setScissorRect:(MTLScissorRect){
        0, 0, (NSUInteger)_viewW, (NSUInteger)_viewH}];
}

void metalDrawQuad(const float* verts) {
    if (!_enc) return;
    [_enc setVertexBytes:verts length:4*36 atIndex:0];
    [_enc drawIndexedPrimitives:MTLPrimitiveTypeTriangle
                     indexCount:6
                      indexType:MTLIndexTypeUInt16
                    indexBuffer:_quadIdx
              indexBufferOffset:0];
}

void metalDrawTriangles(const float* verts, int numVerts) {
    if (!_enc || numVerts <= 0) return;
    int byteLen = numVerts * 36;
    if (byteLen <= 4096) {
        [_enc setVertexBytes:verts length:byteLen atIndex:0];
    } else {
        id<MTLBuffer> buf = nil;
        if (_triBufFrame >= 0 && _triBufFrame < TRI_BUF_RING) {
            int slot = _triBufCursor[_triBufFrame]++;
            if (slot < TRI_BUF_MAX_PER_FRAME) {
                buf = _triBufs[_triBufFrame][slot];
                if (!buf || [buf length] < (NSUInteger)byteLen) {
                    NSUInteger cap = (NSUInteger)byteLen;
                    NSUInteger page = 4096;
                    cap = ((cap + page - 1) / page) * page;
                    buf = [_device newBufferWithLength:cap
                                               options:MTLResourceStorageModeShared];
                    _triBufs[_triBufFrame][slot] = buf;
                }
                memcpy([buf contents], verts, (size_t)byteLen);
            }
        }
        if (!buf) {
            // Pool exhausted — allocate a one-off buffer.
            buf = [_device newBufferWithBytes:verts
                                       length:(NSUInteger)byteLen
                                      options:MTLResourceStorageModeShared];
        }
        if (!buf) return;
        [_enc setVertexBuffer:buf offset:0 atIndex:0];
    }
    [_enc drawPrimitives:MTLPrimitiveTypeTriangle
             vertexStart:0 vertexCount:numVerts];
}

void metalDrawGlyphQuad(const float* verts) {
    if (!_enc) return;
    [_enc setVertexBytes:verts length:4*32 atIndex:0];
    [_enc drawIndexedPrimitives:MTLPrimitiveTypeTriangle
                     indexCount:6
                      indexType:MTLIndexTypeUInt16
                    indexBuffer:_quadIdx
              indexBufferOffset:0];
}

// ─── Textures ─────────────────────────────────────────────────

int metalCreateTexture(int w, int h, const void* pixels,
                       int hasData) {
    int tid = 0;
    if (_freeTexCount > 0) {
        tid = _freeTexIDs[--_freeTexCount];
    } else {
        if (_nextTexID >= MAX_TEX) return 0;
        tid = _nextTexID++;
    }

    id<MTLTexture> tex = makeTexture(w, h,
        MTLPixelFormatRGBA8Unorm);
    if (!tex) {
        if (_freeTexCount < MAX_TEX) {
            _freeTexIDs[_freeTexCount++] = tid;
        }
        return 0;
    }
    if (hasData && pixels) {
        [tex replaceRegion:MTLRegionMake2D(0, 0, w, h)
               mipmapLevel:0
                 withBytes:pixels
               bytesPerRow:w * 4];
    }
    _textures[tid] = tex;
    return tid;
}

void metalUpdateTexture(int id, int x, int y, int w, int h,
                        const void* data) {
    if (id <= 0 || id >= MAX_TEX || !_textures[id]) return;
    [_textures[id] replaceRegion:MTLRegionMake2D(x, y, w, h)
                     mipmapLevel:0
                       withBytes:data
                     bytesPerRow:w * 4];
}

void metalDeleteTexture(int id) {
    if (id <= 0 || id >= MAX_TEX) return;
    if (!_textures[id]) return;
    _textures[id] = nil;
    if (_freeTexCount < MAX_TEX) {
        _freeTexIDs[_freeTexCount++] = id;
    }
}

void metalBindTexture(int id) {
    if (!_enc) return;
    if (id > 0 && id < MAX_TEX && _textures[id]) {
        [_enc setFragmentTexture:_textures[id] atIndex:0];
    }
}

// ─── Filter System ────────────────────────────────────────────

static void ensureFilterTextures(int w, int h) {
    if (_filterTexA && _filterW == w && _filterH == h) return;
    MTLPixelFormat pf = MTLPixelFormatBGRA8Unorm;
    _filterTexA = makeRenderTarget(w, h, pf);
    _filterTexB = makeRenderTarget(w, h, pf);
    _filterW = w;
    _filterH = h;
}

int metalBeginFilter(int w, int h) {
    if (!_enc || !_cmdBuf) return -1;

    ensureFilterTextures(w, h);
    if (!_filterTexA || !_filterTexB) return -2;

    // End current main encoder.
    [_enc endEncoding];
    _enc = nil;

    // Start render pass targeting filterTexA.
    MTLRenderPassDescriptor *rpd =
        [MTLRenderPassDescriptor renderPassDescriptor];
    rpd.colorAttachments[0].texture     = _filterTexA;
    rpd.colorAttachments[0].loadAction  = MTLLoadActionClear;
    rpd.colorAttachments[0].storeAction = MTLStoreActionStore;
    rpd.colorAttachments[0].clearColor  =
        MTLClearColorMake(0, 0, 0, 0);

    _enc = [_cmdBuf renderCommandEncoderWithDescriptor:rpd];
    [_enc setViewport:(MTLViewport){
        0, 0, (double)w, (double)h, 0, 1}];
    [_enc setFragmentSamplerState:_sampler atIndex:0];
    return 0;
}

void metalEndFilter(float blurRadius, int layers) {
    if (!_enc || !_cmdBuf) return;

    float stdDev = blurRadius;
    if (stdDev < 1) stdDev = 1;
    if (layers < 1) layers = 1;

    int w = _filterW;
    int h = _filterH;

    // End filter content encoder.
    [_enc endEncoding];
    _enc = nil;

    // ── Horizontal blur: filterTexA → filterTexB ──
    {
        MTLRenderPassDescriptor *rpd =
            [MTLRenderPassDescriptor renderPassDescriptor];
        rpd.colorAttachments[0].texture     = _filterTexB;
        rpd.colorAttachments[0].loadAction  = MTLLoadActionClear;
        rpd.colorAttachments[0].storeAction = MTLStoreActionStore;
        rpd.colorAttachments[0].clearColor  =
            MTLClearColorMake(0, 0, 0, 0);

        id<MTLRenderCommandEncoder> enc =
            [_cmdBuf renderCommandEncoderWithDescriptor:rpd];
        [enc setViewport:(MTLViewport){
            0, 0, (double)w, (double)h, 0, 1}];
        [enc setRenderPipelineState:_pipelines[PIPE_FILTER_BLUR_H]];
        [enc setFragmentSamplerState:_sampler atIndex:0];
        [enc setFragmentTexture:_filterTexA atIndex:0];

        // TM with stdDev.
        float tm[16] = {0};
        tm[0] = stdDev;
        [enc setVertexBytes:tm length:64 atIndex:2];

        // MVP: identity maps pixel coords to NDC.
        float mvp[16] = {0};
        mvp[0]  =  2.0f / w;
        mvp[5]  = -2.0f / h;
        mvp[10] = -1.0f;
        mvp[12] = -1.0f;
        mvp[13] =  1.0f;
        mvp[15] =  1.0f;
        [enc setVertexBytes:mvp length:64 atIndex:1];

        // Full-screen quad (with texture UVs 0..1).
        float verts[] = {
            0,0,0, 0,1, 1,1,1,1,
            (float)w,0,0, 1,1, 1,1,1,1,
            (float)w,(float)h,0, 1,0, 1,1,1,1,
            0,(float)h,0, 0,0, 1,1,1,1,
        };
        [enc setVertexBytes:verts length:sizeof(verts)
                    atIndex:0];
        [enc drawIndexedPrimitives:MTLPrimitiveTypeTriangle
                        indexCount:6
                         indexType:MTLIndexTypeUInt16
                       indexBuffer:_quadIdx
                 indexBufferOffset:0];
        [enc endEncoding];
    }

    // ── Vertical blur: filterTexB → filterTexA ──
    {
        MTLRenderPassDescriptor *rpd =
            [MTLRenderPassDescriptor renderPassDescriptor];
        rpd.colorAttachments[0].texture     = _filterTexA;
        rpd.colorAttachments[0].loadAction  = MTLLoadActionClear;
        rpd.colorAttachments[0].storeAction = MTLStoreActionStore;
        rpd.colorAttachments[0].clearColor  =
            MTLClearColorMake(0, 0, 0, 0);

        id<MTLRenderCommandEncoder> enc =
            [_cmdBuf renderCommandEncoderWithDescriptor:rpd];
        [enc setViewport:(MTLViewport){
            0, 0, (double)w, (double)h, 0, 1}];
        [enc setRenderPipelineState:_pipelines[PIPE_FILTER_BLUR_V]];
        [enc setFragmentSamplerState:_sampler atIndex:0];
        [enc setFragmentTexture:_filterTexB atIndex:0];

        float tm[16] = {0};
        tm[0] = stdDev;
        [enc setVertexBytes:tm length:64 atIndex:2];

        float mvp[16] = {0};
        mvp[0]  =  2.0f / w;
        mvp[5]  = -2.0f / h;
        mvp[10] = -1.0f;
        mvp[12] = -1.0f;
        mvp[13] =  1.0f;
        mvp[15] =  1.0f;
        [enc setVertexBytes:mvp length:64 atIndex:1];

        float verts[] = {
            0,0,0, 0,1, 1,1,1,1,
            (float)w,0,0, 1,1, 1,1,1,1,
            (float)w,(float)h,0, 1,0, 1,1,1,1,
            0,(float)h,0, 0,0, 1,1,1,1,
        };
        [enc setVertexBytes:verts length:sizeof(verts)
                    atIndex:0];
        [enc drawIndexedPrimitives:MTLPrimitiveTypeTriangle
                        indexCount:6
                         indexType:MTLIndexTypeUInt16
                       indexBuffer:_quadIdx
                 indexBufferOffset:0];
        [enc endEncoding];
    }

    // ── Resume main render pass (load, not clear) ──
    beginMainEncoder(0, 0, 0, 0, 0);

    // ── Composite: draw filterTexA onto main drawable ──
    [_enc setRenderPipelineState:_pipelines[PIPE_FILTER_TEX]];
    [_enc setFragmentTexture:_filterTexA atIndex:0];

    float mvp[16] = {0};
    mvp[0]  =  2.0f / _viewW;
    mvp[5]  = -2.0f / _viewH;
    mvp[10] = -1.0f;
    mvp[12] = -1.0f;
    mvp[13] =  1.0f;
    mvp[15] =  1.0f;
    [_enc setVertexBytes:mvp length:64 atIndex:1];

    float tm[16] = {0};
    tm[0] = 1; tm[5] = 1; tm[10] = 1; tm[15] = 1;
    [_enc setVertexBytes:tm length:64 atIndex:2];

    float glowAlpha = (float)(60 * layers);
    if (glowAlpha > 255) glowAlpha = 255;
    glowAlpha /= 255.0f;
    float a = glowAlpha;

    float verts[] = {
        0,0,0, 0,1, 1,1,1,a,
        (float)_viewW,0,0, 1,1, 1,1,1,a,
        (float)_viewW,(float)_viewH,0, 1,0, 1,1,1,a,
        0,(float)_viewH,0, 0,0, 1,1,1,a,
    };
    [_enc setVertexBytes:verts length:sizeof(verts) atIndex:0];
    [_enc drawIndexedPrimitives:MTLPrimitiveTypeTriangle
                     indexCount:6
                      indexType:MTLIndexTypeUInt16
                    indexBuffer:_quadIdx
              indexBufferOffset:0];
}
