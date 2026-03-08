package gui

// Axis defines if a Layout arranges children horizontally,
// vertically, or not at all.
type Axis uint8

// Axis constants.
const (
	AxisNone        Axis = iota
	AxisTopToBottom      // vertical
	AxisLeftToRight      // horizontal
)

// HorizontalAlign specifies horizontal alignment.
type HorizontalAlign uint8

// HorizontalAlign constants.
const (
	HAlignStart HorizontalAlign = iota // culture-dependent
	HAlignEnd                          // culture-dependent
	HAlignCenter
	HAlignLeft  // always left
	HAlignRight // always right
)

// VerticalAlign specifies vertical alignment.
type VerticalAlign uint8

// VerticalAlign constants.
const (
	VAlignTop VerticalAlign = iota
	VAlignMiddle
	VAlignBottom
)

// TextAlignment specifies horizontal text alignment.
type TextAlignment uint8

// TextAlignment constants.
const (
	TextAlignLeft TextAlignment = iota
	TextAlignCenter
	TextAlignRight
)
