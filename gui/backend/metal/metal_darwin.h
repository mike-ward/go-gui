#ifndef METAL_DARWIN_H
#define METAL_DARWIN_H

#include <stdint.h>

// Pipeline IDs — must match Go constants.
enum {
    PIPE_SOLID = 0,
    PIPE_SHADOW,
    PIPE_BLUR,
    PIPE_GRADIENT,
    PIPE_IMAGE_CLIP,
    PIPE_FILTER_BLUR_H,
    PIPE_FILTER_BLUR_V,
    PIPE_FILTER_TEX,
    PIPE_FILTER_COLOR,
    PIPE_GLYPH_TEX,
    PIPE_GLYPH_COLOR,
    PIPE_STENCIL,
    PIPE_COUNT,
};

// Opaque context handle — one per window. Internally a
// MetalContext* defined in the .m file.
typedef void* MetalCtx;

// Lifecycle
MetalCtx metalCtxCreate(void* metalLayer);
void metalCtxDestroy(MetalCtx ctx);
void metalResize(MetalCtx ctx, int w, int h);

// Frame
int  metalBeginFrame(MetalCtx ctx,
                     float r, float g, float b, float a);
void metalEndFrame(MetalCtx ctx);

// Pipeline and uniforms
void metalSetPipeline(MetalCtx ctx, int id);
void metalSetMVP(MetalCtx ctx, const float* m);
void metalSetTM(MetalCtx ctx, const float* m);

// Scissor
void metalSetScissor(MetalCtx ctx,
                     int x, int y, int w, int h, int viewH);
void metalDisableScissor(MetalCtx ctx);

// Drawing (main vertex format: 9 floats/vertex)
void metalDrawQuad(MetalCtx ctx, const float* verts);
void metalDrawTriangles(MetalCtx ctx,
                        const float* verts, int numVerts);

// Drawing (glyph vertex format: 8 floats/vertex)
void metalDrawGlyphQuad(MetalCtx ctx, const float* verts);

// Textures
int  metalCreateTexture(MetalCtx ctx,
                        int w, int h, const void* pixels,
                        int hasData);
void metalUpdateTexture(MetalCtx ctx,
                        int id, int x, int y, int w, int h,
                        const void* data);
void metalDeleteTexture(MetalCtx ctx, int id);
void metalBindTexture(MetalCtx ctx, int id);

// Custom shader pipelines
int  metalBuildCustomPipeline(MetalCtx ctx,
                              const char* mslSrc);
void metalDeleteCustomPipeline(MetalCtx ctx, int idx);
void metalSetCustomPipeline(MetalCtx ctx, int idx);

// Filter (glow) system
int  metalBeginFilter(MetalCtx ctx, int w, int h);
void metalEndFilter(MetalCtx ctx, float blurRadius,
                    int layers, const float* colorMatrix);

// Stencil clip
void metalBeginStencilClip(MetalCtx ctx,
                           const float* verts, int depth);
void metalEndStencilClip(MetalCtx ctx,
                         const float* verts, int depth);

#endif
