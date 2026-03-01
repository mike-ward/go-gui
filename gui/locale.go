package gui

// Locale holds locale-specific settings.
type Locale struct {
	TextDir TextDirection
}

// guiLocale is the global locale setting.
var guiLocale = Locale{TextDir: TextDirLTR}

// effectiveTextDir resolves the text direction for a shape,
// falling back to the global locale when set to Auto.
func effectiveTextDir(shape *Shape) TextDirection {
	if shape.TextDir != TextDirAuto {
		return shape.TextDir
	}
	return guiLocale.TextDir
}
