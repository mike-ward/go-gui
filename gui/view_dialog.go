package gui

// DialogType identifies the kind of dialog.
type DialogType uint8

// DialogType constants.
const (
	DialogMessage DialogType = iota
	DialogConfirm
	DialogPrompt
	DialogCustom
)

const dialogBaseIDFocus uint32 = 7568971

// DialogCfg configures a modal dialog.
type DialogCfg struct {
	Title       string
	Body        string
	Reply       string
	ID          string
	Color       Color
	ColorBorder Color
	Padding     Opt[Padding]
	SizeBorder  Opt[float32]

	TitleTextStyle TextStyle
	TextStyle      TextStyle

	CustomContent []View

	OnOkYes    func(*Window)
	OnCancelNo func(*Window)
	OnReply    func(string, *Window)

	Width     float32
	Height    float32
	MinWidth  float32
	MinHeight float32
	MaxWidth  float32
	MaxHeight float32

	Radius       Opt[float32]
	RadiusBorder Opt[float32]

	IDFocus      uint32
	DialogType   DialogType
	AlignButtons HorizontalAlign

	// unexported
	visible    bool
	oldIDFocus uint32
}

// dialogViewGenerator builds the dialog overlay view from cfg.
func dialogViewGenerator(cfg DialogCfg) View {
	applyDialogDefaults(&cfg)
	dn := &DefaultDialogStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)
	_ = cfg.RadiusBorder.Get(dn.RadiusBorder)

	var content []View

	// Title.
	if cfg.Title != "" {
		content = append(content, Text(TextCfg{
			Text:      cfg.Title,
			TextStyle: cfg.TitleTextStyle,
		}))
	}

	// Body (unless custom).
	if cfg.DialogType != DialogCustom && cfg.Body != "" {
		content = append(content, Text(TextCfg{
			Text:      cfg.Body,
			TextStyle: cfg.TextStyle,
			Mode:      TextModeWrap,
		}))
	}

	// Type-specific content.
	switch cfg.DialogType {
	case DialogMessage:
		content = append(content, messageView(cfg))
	case DialogConfirm:
		content = append(content, confirmView(cfg))
	case DialogPrompt:
		content = append(content, promptView(cfg)...)
	case DialogCustom:
		content = append(content, cfg.CustomContent...)
	}

	return Column(ContainerCfg{
		ID:          reservedDialogID,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(sizeBorder),
		Radius:      Some(radius),
		Padding:     cfg.Padding,
		Width:       cfg.Width,
		Height:      cfg.Height,
		MinWidth:    cfg.MinWidth,
		MinHeight:   cfg.MinHeight,
		MaxWidth:    cfg.MaxWidth,
		MaxHeight:   cfg.MaxHeight,
		Float:       true,
		FloatAnchor: FloatMiddleCenter,
		FloatTieOff: FloatMiddleCenter,
		Spacing:     Some(SpacingMedium),
		OnKeyDown:   dialogKeyDown(cfg),
		A11YRole:    AccessRoleDialog,
		A11YState:   AccessStateModal,
		Content:     content,
	})
}

// messageView returns an OK button row.
func messageView(cfg DialogCfg) View {
	onOkYes := cfg.OnOkYes
	oldFocus := cfg.oldIDFocus
	return Row(ContainerCfg{
		Sizing:  FillFit,
		HAlign:  cfg.AlignButtons,
		Padding: Some(PaddingNone),
		Content: []View{
			Button(ButtonCfg{
				IDFocus: cfg.IDFocus,
				Content: []View{Text(TextCfg{Text: "OK"})},
				OnClick: func(_ *Layout, _ *Event, w *Window) {
					w.SetIDFocus(oldFocus)
					w.DialogDismiss()
					if onOkYes != nil {
						onOkYes(w)
					}
				},
			}),
		},
	})
}

// confirmView returns Yes/No button row.
func confirmView(cfg DialogCfg) View {
	onOkYes := cfg.OnOkYes
	onCancelNo := cfg.OnCancelNo
	oldFocus := cfg.oldIDFocus
	return Row(ContainerCfg{
		Sizing:  FillFit,
		HAlign:  cfg.AlignButtons,
		Padding: Some(PaddingNone),
		Spacing: Some(SpacingMedium),
		Content: []View{
			Button(ButtonCfg{
				IDFocus: cfg.IDFocus + 1,
				Content: []View{Text(TextCfg{Text: "Yes"})},
				OnClick: func(_ *Layout, _ *Event, w *Window) {
					w.SetIDFocus(oldFocus)
					w.DialogDismiss()
					if onOkYes != nil {
						onOkYes(w)
					}
				},
			}),
			Button(ButtonCfg{
				IDFocus: cfg.IDFocus,
				Content: []View{Text(TextCfg{Text: "No"})},
				OnClick: func(_ *Layout, _ *Event, w *Window) {
					w.SetIDFocus(oldFocus)
					w.DialogDismiss()
					if onCancelNo != nil {
						onCancelNo(w)
					}
				},
			}),
		},
	})
}

// promptView returns input + OK/Cancel button row.
func promptView(cfg DialogCfg) []View {
	onReply := cfg.OnReply
	onCancelNo := cfg.OnCancelNo
	oldFocus := cfg.oldIDFocus

	var views []View

	views = append(views, Input(InputCfg{
		ID:      "dialog_prompt_input",
		Text:    cfg.Reply,
		Sizing:  FillFit,
		IDFocus: cfg.IDFocus,
		OnTextChanged: func(_ *Layout, text string, w *Window) {
			w.dialogCfg.Reply = text
		},
	}))

	views = append(views, Row(ContainerCfg{
		Sizing:  FillFit,
		HAlign:  cfg.AlignButtons,
		Padding: Some(PaddingNone),
		Spacing: Some(SpacingMedium),
		Content: []View{
			Button(ButtonCfg{
				IDFocus:  cfg.IDFocus + 1,
				Disabled: len(cfg.Reply) == 0,
				Content:  []View{Text(TextCfg{Text: "OK"})},
				OnClick: func(_ *Layout, _ *Event, w *Window) {
					w.SetIDFocus(oldFocus)
					reply := w.dialogCfg.Reply
					w.DialogDismiss()
					if onReply != nil {
						onReply(reply, w)
					}
				},
			}),
			Button(ButtonCfg{
				IDFocus: cfg.IDFocus + 2,
				Content: []View{Text(TextCfg{Text: "Cancel"})},
				OnClick: func(_ *Layout, _ *Event, w *Window) {
					w.SetIDFocus(oldFocus)
					w.DialogDismiss()
					if onCancelNo != nil {
						onCancelNo(w)
					}
				},
			}),
		},
	}))

	return views
}

// dialogKeyDown handles Escape to dismiss dialog.
func dialogKeyDown(cfg DialogCfg) func(*Layout, *Event, *Window) {
	onCancelNo := cfg.OnCancelNo
	return func(_ *Layout, e *Event, w *Window) {
		if e.KeyCode == KeyEscape {
			w.DialogDismiss()
			if onCancelNo != nil {
				onCancelNo(w)
			}
			e.IsHandled = true
			return
		}
		if e.KeyCode == KeyC &&
			e.Modifiers.HasAny(ModCtrl, ModSuper) &&
			cfg.Body != "" {
			w.SetClipboard(cfg.Body)
			e.IsHandled = true
		}
	}
}

func applyDialogDefaults(cfg *DialogCfg) {
	d := &DefaultDialogStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}
	if cfg.TitleTextStyle == (TextStyle{}) {
		cfg.TitleTextStyle = d.TitleTextStyle
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyle
	}
	if cfg.IDFocus == 0 {
		cfg.IDFocus = dialogBaseIDFocus
	}
	if cfg.AlignButtons == HAlignStart {
		cfg.AlignButtons = d.AlignButtons
	}
	if cfg.MinWidth == 0 {
		cfg.MinWidth = 200
	}
	if cfg.MaxWidth == 0 {
		cfg.MaxWidth = 300
	}
}

// Dialog shows a modal dialog.
func (w *Window) Dialog(cfg DialogCfg) {
	applyDialogDefaults(&cfg)
	cfg.visible = true
	cfg.oldIDFocus = w.viewState.idFocus
	w.dialogCfg = cfg
	w.SetIDFocus(cfg.IDFocus)
}

// DialogDismiss closes the current dialog.
func (w *Window) DialogDismiss() {
	oldFocus := w.dialogCfg.oldIDFocus
	w.dialogCfg = DialogCfg{}
	w.SetIDFocus(oldFocus)
}

// DialogIsVisible returns true if a dialog is showing.
func (w *Window) DialogIsVisible() bool {
	return w.dialogCfg.visible
}
