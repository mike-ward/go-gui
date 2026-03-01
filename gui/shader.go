package gui

import (
	"runtime"
	"strings"
)

// Shader holds custom fragment shader bodies and parameters.
type Shader struct {
	Metal  string    // MSL fragment body
	GLSL   string    // GLSL 3.3 fragment body
	Params []float32 // up to 16 custom floats
}

// BuildMetalFragment wraps a user-supplied MSL body with the
// standard preamble (struct defs, SDF clipping) and epilogue.
func BuildMetalFragment(body string) string {
	var b strings.Builder
	b.WriteString(`
#include <metal_stdlib>
using namespace metal;

struct VertexOut {
    float4 position [[position]];
    float2 uv;
    float4 color;
    float params;
    float4 p0;
    float4 p1;
    float4 p2;
    float4 p3;
};

fragment float4 fs_main(VertexOut in [[stage_in]], texture2d<float> tex [[texture(0)]], sampler smp [[sampler(0)]]) {
    float radius = floor(in.params / 4096.0) / 4.0;

    float2 width_inv = float2(fwidth(in.uv.x), fwidth(in.uv.y));
    float2 half_size = 1.0 / (width_inv + 1e-6);
    float2 pos = in.uv * half_size;

    float2 q = abs(pos) - half_size + float2(radius);
    float2 max_q = max(q, float2(0.0));
    float d = length(max_q) + min(max(q.x, q.y), 0.0) - radius;

    float grad_len = length(float2(dfdx(d), dfdy(d)));
    d = d / max(grad_len, 0.001);
    float sdf_alpha = 1.0 - smoothstep(-0.59, 0.59, d);

    // --- user body ---
    `)
	b.WriteString(body)
	b.WriteString(`
    // --- end user body ---

    frag_color = float4(frag_color.rgb, frag_color.a * sdf_alpha);

    if (frag_color.a < 0.0) {
        frag_color += tex.sample(smp, in.uv);
    }
    return frag_color;
}
`)
	return b.String()
}

// BuildGLSLFragment wraps a user-supplied GLSL body with the
// standard preamble and epilogue.
func BuildGLSLFragment(body string) string {
	var b strings.Builder
	b.WriteString(`
    #version 330
    uniform sampler2D tex;
    in vec2 uv;
    in vec4 color;
    in float params;
    in vec4 p0;
    in vec4 p1;
    in vec4 p2;
    in vec4 p3;

    out vec4 _frag_out;

    void main() {
        float radius = floor(params / 4096.0) / 4.0;

        vec2 uv_to_px = 1.0 / (vec2(fwidth(uv.x), fwidth(uv.y)) + 1e-6);
        vec2 half_size = uv_to_px;
        vec2 pos = uv * half_size;

        vec2 q = abs(pos) - half_size + vec2(radius);
        float d = length(max(q, 0.0)) + min(max(q.x, q.y), 0.0) - radius;

        float grad_len = length(vec2(dFdx(d), dFdy(d)));
        d = d / max(grad_len, 0.001);
        float sdf_alpha = 1.0 - smoothstep(-0.59, 0.59, d);

        // --- user body ---
        `)
	b.WriteString(body)
	b.WriteString(`
        // --- end user body ---

        _frag_out = vec4(frag_color.rgb, frag_color.a * sdf_alpha);

        if (_frag_out.a < 0.0) {
            _frag_out += texture(tex, uv);
        }
    }
`)
	return b.String()
}

// ShaderHash computes a cache key from the shader source.
// Uses Metal source on macOS, GLSL otherwise.
func ShaderHash(s *Shader) uint64 {
	if runtime.GOOS == "darwin" {
		return hashString(s.Metal)
	}
	return hashString(s.GLSL)
}

// hashString computes a 64-bit FNV-1a hash.
func hashString(s string) uint64 {
	h := uint64(0xcbf29ce484222325)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 0x100000001b3
	}
	return h
}
