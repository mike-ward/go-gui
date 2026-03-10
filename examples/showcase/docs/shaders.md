# Custom Shaders

go-gui supports custom fragment shaders on containers and rectangles.
Write only the color computation body — the framework handles struct
definitions, SDF round-rect clipping, and pipeline caching.

## API Reference

### Shader

```go
type Shader struct {
    Metal  string    // MSL fragment body
    GLSL   string    // GLSL 3.3 fragment body
    Params []float32 // up to 16 custom floats
}
```

### Available Inputs

User shader body has access to these variables:

| Metal          | GLSL          | Type        | Description            |
|----------------|---------------|-------------|------------------------|
| in.uv          | uv            | float2/vec2 | -1..1 centered coords |
| in.color       | color         | float4/vec4 | vertex color           |
| in.params      | params        | float       | packed radius          |
| in.p0..in.p3   | p0..p3        | float4/vec4 | custom params          |
| in.position    | gl_FragCoord  | float4/vec4 | screen position        |

### Output

Declare a local `float4 frag_color` (Metal) or `vec4 frag_color`
(GLSL) with the desired RGBA color. The framework applies SDF
clipping and the dummy texture workaround automatically.

## Usage

### Basic Shader on a Container

```go
gui.Column(gui.ContainerCfg{
    Width:  200,
    Height: 200,
    Sizing: gui.FixedFixed,
    Radius: gui.SomeF(8),
    Shader: &gui.Shader{
        Metal: `
            float2 st = in.uv * 0.5 + 0.5;
            float4 frag_color = float4(st.x, st.y, 0.5, 1.0);
        `,
        GLSL: `
            vec2 st = uv * 0.5 + 0.5;
            vec4 frag_color = vec4(st.x, st.y, 0.5, 1.0);
        `,
    },
    Content: []gui.View{...},
})
```

### Animated Shader with Params

Pass time or other values via `Params`. Each float maps to
`p0.x`, `p0.y`, `p0.z`, `p0.w`, `p1.x`, etc.

```go
elapsed := float32(time.Since(startTime).Milliseconds()) / 1000.0

gui.Column(gui.ContainerCfg{
    Width:  200,
    Height: 200,
    Sizing: gui.FixedFixed,
    Radius: gui.SomeF(16),
    Shader: &gui.Shader{
        Metal: `
            float t = in.p0.x;
            float2 st = in.uv * 0.5 + 0.5;
            float3 c = 0.5 + 0.5 * cos(t + st.xyx + float3(0,2,4));
            float4 frag_color = float4(c, 1.0);
        `,
        GLSL: `
            float t = p0.x;
            vec2 st = uv * 0.5 + 0.5;
            vec3 c = 0.5 + 0.5 * cos(t + st.xyx + vec3(0,2,4));
            vec4 frag_color = vec4(c, 1.0);
        `,
        Params: []float32{elapsed},
    },
    Content: []gui.View{...},
})
```

### Shader on a Rectangle

```go
gui.Rectangle(gui.RectangleCfg{
    Width:  100,
    Height: 100,
    Sizing: gui.FixedFixed,
    Radius: gui.SomeF(8),
    Shader: &gui.Shader{
        Metal: `
            float4 frag_color = float4(1.0, 0.0, 0.5, 1.0);
        `,
        GLSL: `
            vec4 frag_color = vec4(1.0, 0.0, 0.5, 1.0);
        `,
    },
})
```

### Shader with Border

Shader replaces the fill only. Borders draw separately:

```go
gui.Column(gui.ContainerCfg{
    Width:       200,
    Height:      200,
    Sizing:      gui.FixedFixed,
    Radius:      gui.SomeF(12),
    SizeBorder:  gui.SomeF(2),
    ColorBorder: gui.White,
    Shader: &gui.Shader{
        Metal: `
            float4 frag_color = float4(0.1, 0.1, 0.3, 1.0);
        `,
        GLSL: `
            vec4 frag_color = vec4(0.1, 0.1, 0.3, 1.0);
        `,
    },
    Content: []gui.View{...},
})
```

## How It Works

### Pipeline Caching

Each unique shader source compiles to a GPU pipeline once. The
framework hashes the active platform's source (Metal on macOS,
GLSL elsewhere) and caches the pipeline. Multiple views sharing
the same shader source reuse one pipeline.

### Rendering Priority

Shader has the highest fill priority:

1. **Shader** — custom fragment shader fill
2. **Gradient** — linear/radial gradient fill
3. **Blur** — blurred rectangle
4. **Rectangle** — solid color fill

Shadows always render before the fill regardless of type.

### Body Wrapping

The framework wraps the user body with:
- Struct definitions (VertexOut with uv, color, params, p0..p3)
- SDF round-rect distance calculation
- Alpha clipping via `smoothstep`
- Dummy texture sample (sokol workaround)

### Params Layout

The 16-float `Params` slice maps to four vec4 varyings:

| Index | Varying | Component   |
|-------|---------|-------------|
| 0-3   | p0      | x, y, z, w |
| 4-7   | p1      | x, y, z, w |
| 8-11  | p2      | x, y, z, w |
| 12-15 | p3      | x, y, z, w |

## Animation

To animate shader parameters, add a repeating animation to keep
the frame loop hot, then update `Params` each frame:

```go
w.AnimationAdd(&gui.Animate{
    AnimateID: "shader_tick",
    Repeat:    true,
    Callback:  func(_ *gui.Animate, _ *gui.Window) {},
})
```

## Limitations

- Must provide both `Metal` and `GLSL` bodies for cross-platform
- No helper function definitions in user body (single expression
  block)
- Maximum 16 float parameters
- Only available on `Column`, `Row`, `Canvas`, `Circle`, and
  `Rectangle` views

## Demo

```bash
go run ./examples/custom_shader/
```
