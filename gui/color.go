package gui

import "fmt"

// Color represents a 32-bit color value in sRGB format.
type Color struct {
	R, G, B, A uint8
}

// Predefined colors.
var (
	Black          = Color{0, 0, 0, 255}
	Gray           = Color{128, 128, 128, 255}
	White          = Color{255, 255, 255, 255}
	Red            = Color{255, 0, 0, 255}
	Green          = Color{0, 255, 0, 255}
	Blue           = Color{0, 0, 255, 255}
	Yellow         = Color{255, 255, 0, 255}
	Magenta        = Color{255, 0, 255, 255}
	Cyan           = Color{0, 255, 255, 255}
	Orange         = Color{255, 165, 0, 255}
	Purple         = Color{128, 0, 128, 255}
	Indigo         = Color{75, 0, 130, 255}
	Pink           = Color{255, 192, 203, 255}
	Violet         = Color{238, 130, 238, 255}
	DarkBlue       = Color{0, 0, 139, 255}
	DarkGray       = Color{169, 169, 169, 255}
	DarkGreen      = Color{0, 100, 0, 255}
	DarkRed        = Color{139, 0, 0, 255}
	LightBlue      = Color{173, 216, 230, 255}
	LightGray      = Color{211, 211, 211, 255}
	LightGreen     = Color{144, 238, 144, 255}
	LightRed       = Color{255, 204, 203, 255}
	CornflowerBlue = Color{100, 149, 237, 255}
	RoyalBlue      = Color{65, 105, 225, 255}
	ColorTransparent = Color{0, 0, 0, 0}
)

// Hex creates a Color from a hexadecimal integer (0xRRGGBB).
func Hex(color int) Color {
	return Color{
		R: uint8((color >> 16) & 0xFF),
		G: uint8((color >> 8) & 0xFF),
		B: uint8(color & 0xFF),
		A: 255,
	}
}

// RGB builds a Color from r, g, b values. Alpha defaults to 255.
func RGB(r, g, b uint8) Color {
	return Color{r, g, b, 255}
}

// RGBA builds a Color from r, g, b, a values.
func RGBA(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// WithOpacity returns color with alpha multiplied by opacity (0.0–1.0).
func (c Color) WithOpacity(opacity float32) Color {
	return Color{
		R: c.R,
		G: c.G,
		B: c.B,
		A: uint8(float32(c.A) * f32Clamp(opacity, 0, 1)),
	}
}

// Add returns a + b, clamping each channel to 255.
func (a Color) Add(b Color) Color {
	return Color{
		R: clampAdd(a.R, b.R),
		G: clampAdd(a.G, b.G),
		B: clampAdd(a.B, b.B),
		A: clampAdd(a.A, b.A),
	}
}

// Sub returns a - b, clamping each channel to 0.
func (a Color) Sub(b Color) Color {
	aa := a.A
	if b.A > aa {
		aa = b.A
	}
	return Color{
		R: clampSub(a.R, b.R),
		G: clampSub(a.G, b.G),
		B: clampSub(a.B, b.B),
		A: aa,
	}
}

// Over implements Porter-Duff "a over b" compositing.
func (a Color) Over(b Color) Color {
	aa := float32(a.A) / 255
	ab := float32(b.A) / 255
	ar := aa + ab*(1-aa)
	if ar == 0 {
		return Color{}
	}
	rr := (float32(a.R)*aa + float32(b.R)*ab*(1-aa)) / ar
	gr := (float32(a.G)*aa + float32(b.G)*ab*(1-aa)) / ar
	br := (float32(a.B)*aa + float32(b.B)*ab*(1-aa)) / ar
	return Color{
		R: uint8(rr),
		G: uint8(gr),
		B: uint8(br),
		A: uint8(ar * 255),
	}
}

// Eq checks if two colors are equal in every channel.
func (c Color) Eq(c2 Color) bool {
	return c.R == c2.R && c.G == c2.G && c.B == c2.B && c.A == c2.A
}

// String returns a string representation.
func (c Color) String() string {
	return fmt.Sprintf("Color{%d, %d, %d, %d}", c.R, c.G, c.B, c.A)
}

// RGBA8 converts to an int in RGBA8 order.
func (c Color) RGBA8() int {
	return int(uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A))
}

// BGRA8 converts to an int in BGRA8 order.
func (c Color) BGRA8() int {
	return int(uint32(c.B)<<24 | uint32(c.G)<<16 | uint32(c.R)<<8 | uint32(c.A))
}

// ABGR8 converts to an int in ABGR8 order.
func (c Color) ABGR8() int {
	return int(uint32(c.A)<<24 | uint32(c.B)<<16 | uint32(c.G)<<8 | uint32(c.R))
}

// ToCSSString returns CSS-compatible "rgba(r,g,b,a)".
func (c Color) ToCSSString() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
}

var stringColors = map[string]Color{
	"blue":            Blue,
	"red":             Red,
	"green":           Green,
	"yellow":          Yellow,
	"orange":          Orange,
	"purple":          Purple,
	"black":           Black,
	"gray":            Gray,
	"indigo":          Indigo,
	"pink":            Pink,
	"violet":          Violet,
	"white":           White,
	"cornflower_blue": CornflowerBlue,
	"royal_blue":      RoyalBlue,
	"dark_blue":       DarkBlue,
	"dark_gray":       DarkGray,
	"dark_green":      DarkGreen,
	"dark_red":        DarkRed,
	"light_blue":      LightBlue,
	"light_gray":      LightGray,
	"light_green":     LightGreen,
	"light_red":       LightRed,
}

// ColorFromString returns a Color for the given name or "#RRGGBB" hex
// string. Returns Black if not found.
func ColorFromString(s string) Color {
	if len(s) > 0 && s[0] == '#' {
		c, ok := ColorFromHexString(s)
		if ok {
			return c
		}
		return Color{A: 255}
	}
	if c, ok := stringColors[s]; ok {
		return c
	}
	return Color{A: 255}
}

func clampAdd(a, b uint8) uint8 {
	s := int(a) + int(b)
	if s > 255 {
		return 255
	}
	return uint8(s)
}

func clampSub(a, b uint8) uint8 {
	s := int(a) - int(b)
	if s < 0 {
		return 0
	}
	return uint8(s)
}
