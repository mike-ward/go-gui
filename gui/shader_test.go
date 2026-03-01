package gui

import (
	"strings"
	"testing"
)

func TestBuildMetalFragment(t *testing.T) {
	body := "float4 frag_color = in.color;"
	result := BuildMetalFragment(body)
	if !strings.Contains(result, body) {
		t.Error("body not found in output")
	}
	if !strings.Contains(result, "metal_stdlib") {
		t.Error("missing Metal preamble")
	}
	if !strings.Contains(result, "sdf_alpha") {
		t.Error("missing SDF clipping")
	}
}

func TestBuildGLSLFragment(t *testing.T) {
	body := "vec4 frag_color = color;"
	result := BuildGLSLFragment(body)
	if !strings.Contains(result, body) {
		t.Error("body not found in output")
	}
	if !strings.Contains(result, "#version 330") {
		t.Error("missing GLSL version")
	}
}

func TestShaderHash(t *testing.T) {
	s := &Shader{Metal: "test_metal", GLSL: "test_glsl"}
	h := ShaderHash(s)
	if h == 0 {
		t.Error("hash should not be zero")
	}
	// Same input → same hash.
	if ShaderHash(s) != h {
		t.Error("hash should be deterministic")
	}
	// Different input → different hash.
	s2 := &Shader{Metal: "other", GLSL: "other"}
	if ShaderHash(s2) == h {
		t.Error("different shaders should differ")
	}
}
