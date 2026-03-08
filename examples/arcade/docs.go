package main

import "github.com/mike-ward/go-gui/gui"

func demoDoc(w *gui.Window, source string) gui.View {
	return w.Markdown(gui.MarkdownCfg{Source: source})
}

const docGetStarted = `# Getting Started

## Minimal App

` + "```go" + `
package main

import (
    "github.com/mike-ward/go-gui/gui"
    "github.com/mike-ward/go-gui/gui/backend"
)

type App struct{ Count int }

func main() {
    w := gui.NewWindow(gui.WindowCfg{
        State: &App{}, Title: "Counter",
        Width: 400, Height: 300,
        OnInit: func(w *gui.Window) { w.UpdateView(view) },
    })
    backend.Run(w)
}

func view(w *gui.Window) gui.View {
    app := gui.State[App](w)
    return gui.Column(gui.ContainerCfg{
        Sizing: gui.FillFill, HAlign: gui.HAlignCenter,
        VAlign: gui.VAlignMiddle, Spacing: gui.Some(float32(8)),
        Content: []gui.View{
            gui.Text(gui.TextCfg{Text: fmt.Sprintf("Count: %d", app.Count)}),
            gui.Button(gui.ButtonCfg{
                ID: "inc",
                Content: []gui.View{gui.Text(gui.TextCfg{Text: "+"})},
                OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
                    gui.State[App](w).Count++
                    e.IsHandled = true
                },
            }),
        },
    })
}
` + "```" + `

## Key Concepts

- **One state per window**: ` + "`gui.State[T](w)`" + ` returns ` + "`*T`" + `
- **Immediate mode**: the view function runs every frame
- **No virtual DOM**: layout tree is built fresh each frame
- **Event callbacks**: set ` + "`e.IsHandled = true`" + ` to consume
`

const docArchitecture = `# Architecture

## Rendering Pipeline

` + "```" + `
View fn → Layout tree → layoutArrange() → renderLayout() → []RenderCmd → Backend
` + "```" + `

The framework is **immediate-mode** — no virtual DOM, no diffing.
Each frame rebuilds the entire layout tree from the view function.

## State Management

One typed state slot per window. No globals, no closures capturing mutable state.

` + "```go" + `
w := gui.NewWindow(gui.WindowCfg{State: &MyApp{}})
app := gui.State[MyApp](w) // type-asserts; panics if wrong type
` + "```" + `

## Sizing Model

Sizing combines both axes into a single enum:

| Constant | Width | Height |
|----------|-------|--------|
| FitFit | Fit | Fit |
| FillFill | Fill | Fill |
| FixedFixed | Fixed | Fixed |
| FillFit | Fill | Fit |
| FixedFill | Fixed | Fill |

**Fit**: shrink to content. **Fill**: expand to parent. **Fixed**: use Width/Height values.

## Core Types

- **Layout** — tree node with Shape, parent pointer, children
- **Shape** — renderable: position, size, color, type, events
- **RenderCmd** — single draw operation sent to backend
- **View** — interface satisfied by *Layout
`

const docContainers = `# Containers

## Row, Column, Wrap

All containers accept ` + "`ContainerCfg`" + `. The key differences:

| Container | Axis | Behavior |
|-----------|------|----------|
| Row | Horizontal | Children flow left to right |
| Column | Vertical | Children flow top to bottom |
| Wrap | Horizontal | Wraps to next line when full |

## Alignment

- **HAlign**: Left (default), Center, Right
- **VAlign**: Top (default), Middle, Bottom

## Spacing & Padding

` + "```go" + `
gui.Column(gui.ContainerCfg{
    Spacing: gui.Some(float32(8)),          // gap between children
    Padding: gui.Some(gui.NewPadding(16, 16, 16, 16)), // top, right, bottom, left
})
` + "```" + `

## Scrolling

Add ` + "`IDScroll`" + ` and ` + "`ScrollbarCfgY`" + ` to enable vertical scrolling:

` + "```go" + `
gui.Column(gui.ContainerCfg{
    IDScroll:      myScrollID,
    ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: 4},
    Content:       views,
})
` + "```" + `
`

const docThemes = `# Themes

## Built-in Themes

| Theme | Description |
|-------|-------------|
| ThemeDark | Dark, no borders |
| ThemeDarkBordered | Dark with borders |
| ThemeDarkNoPadding | Dark, zero padding |
| ThemeLight | Light, no borders |
| ThemeLightBordered | Light with borders |
| ThemeLightNoPadding | Light, zero padding |

## Switching Themes

` + "```go" + `
gui.SetTheme(gui.ThemeLight)    // global
w.SetTheme(gui.ThemeLight)      // per-window + refresh
` + "```" + `

## Custom Themes

Use ` + "`ThemeMaker`" + ` to build a theme from ` + "`ThemeCfg`" + `:

` + "```go" + `
cfg := gui.ThemeCfg{
    Name:            "custom",
    ColorBackground: gui.ColorFromHSV(220, 0.15, 0.19),
    ColorPanel:      gui.ColorFromHSV(220, 0.15, 0.25),
    // ... other colors ...
    TextStyleDef:    gui.ThemeDarkCfg.TextStyleDef,
    TitlebarDark:    true,
    SizeBorder:      1,
    Radius:          6,
}
theme := gui.ThemeMaker(cfg)
gui.SetTheme(theme)
` + "```" + `

## Text Styles

Each theme provides text style shortcuts:

- **N1–N6**: Normal (extra small → extra large)
- **B1–B6**: Bold
- **I1–I6**: Italic
- **M1–M6**: Monospace
- **Icon1–Icon6**: Icon font sizes
`

const docAnimations = `# Animations

## Tween

Interpolates a value from A to B with easing:

` + "```go" + `
a := gui.NewTweenAnimation("my-tween", fromVal, toVal,
    func(v float32, w *gui.Window) {
        gui.State[App](w).X = v
    })
a.Easing = gui.EaseOutBounce // optional
w.AnimationAdd(a)
` + "```" + `

## Spring

Physics-based spring animation:

` + "```go" + `
a := gui.NewSpringAnimation("my-spring",
    func(v float32, w *gui.Window) {
        gui.State[App](w).X = v
    })
a.SpringTo(currentX, targetX)
w.AnimationAdd(a)
` + "```" + `

## Keyframes

Multi-waypoint animation with per-segment easing:

` + "```go" + `
a := gui.NewKeyframeAnimation("my-kf",
    []gui.Keyframe{
        {At: 0, Value: 0, Easing: gui.EaseLinear},
        {At: 0.5, Value: 300, Easing: gui.EaseOutCubic},
        {At: 1.0, Value: 0, Easing: gui.EaseOutBounce},
    },
    func(v float32, w *gui.Window) {
        gui.State[App](w).X = v
    })
w.AnimationAdd(a)
` + "```" + `

## Available Easings

EaseLinear, EaseInQuad, EaseOutQuad, EaseInOutQuad,
EaseInCubic, EaseOutCubic, EaseInOutCubic,
EaseInBack, EaseOutBack, EaseOutElastic, EaseOutBounce
`

const docLocales = `# Locales

## Built-in Locales

en-US, de-DE, fr-FR, es-ES, pt-BR, ja-JP, zh-CN, ko-KR, ar-SA, he-IL

## Switching

` + "```go" + `
w.SetLocaleID("de-DE")       // by registry ID
w.SetLocale(myLocale)        // custom locale struct
gui.LocaleAutoDetect()       // detect from OS
` + "```" + `

## Formatting

` + "```go" + `
locale := gui.CurrentLocale()
short := gui.LocaleFormatDate(time.Now(), locale.Date.ShortDate)
long  := gui.LocaleFormatDate(time.Now(), locale.Date.LongDate)
` + "```" + `

## Locale Struct

Each locale defines:
- **Number**: decimal separator, grouping separator, group sizes
- **Date**: short/long format strings, first day of week, 24h clock
- **Currency**: symbol, position, decimal places
- **Strings**: OK, Yes, No, Cancel, Save, Delete, etc.

## Custom Locales

` + "```go" + `
locale, err := gui.LocaleParse(jsonString) // from JSON
gui.LocaleRegister(locale)                 // add to registry
` + "```" + `
`
