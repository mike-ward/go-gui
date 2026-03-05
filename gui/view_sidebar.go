package gui

import "time"

// SidebarRuntimeState tracks animation state for a sidebar.
type SidebarRuntimeState struct {
	PrevOpen    bool
	AnimFrac    float32
	Initialized bool
}

// SidebarCfg configures a sidebar view.
type SidebarCfg struct {
	ID      string
	Open    bool
	Width   float32
	Content []View
	Sizing  Sizing
	Color   Color
	Shadow  *BoxShadow
	Radius  float32
	Padding Padding
	Clip    bool
	// TweenDuration > 0 uses tween; 0 uses spring.
	Spring        SpringCfg
	TweenDuration time.Duration
	TweenEasing   EasingFn
	Disabled      bool
	Invisible     bool

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// Sidebar creates an animated panel that slides in/out.
func (w *Window) Sidebar(cfg SidebarCfg) View {
	if cfg.Width == 0 {
		cfg.Width = 250
	}
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FixedFill
	}
	if !cfg.Color.IsSet() {
		cfg.Color = guiTheme.ColorPanel
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = guiTheme.ContainerStyle.Padding
	}
	if cfg.Spring == (SpringCfg{}) {
		cfg.Spring = SpringStiff
	}
	if cfg.TweenDuration == 0 && cfg.TweenEasing == nil {
		cfg.TweenDuration = 300 * time.Millisecond
		cfg.TweenEasing = EaseInOutCubic
	}

	if cfg.Invisible {
		return invisibleContainerView()
	}

	animW := sidebarAnimatedWidth(w, cfg)
	padW := cfg.Padding.Left + cfg.Padding.Right
	pad := cfg.Padding
	if animW <= padW {
		pad = Padding{}
	}

	return Column(ContainerCfg{
		ID:              cfg.ID,
		Sizing:          cfg.Sizing,
		Width:           animW,
		Padding:         pad,
		Color:           cfg.Color,
		Shadow:          cfg.Shadow,
		Radius:          Some(cfg.Radius),
		Clip:            cfg.Clip,
		Disabled:        cfg.Disabled,
		A11YRole:        AccessRoleGroup,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.ID),
		A11YDescription: cfg.A11YDescription,
		Content:         cfg.Content,
	})
}

func sidebarAnimatedWidth(w *Window, cfg SidebarCfg) float32 {
	sm := StateMap[string, SidebarRuntimeState](
		w, nsSidebar, capFew)

	rt, ok := sm.Get(cfg.ID)
	if !ok {
		rt = SidebarRuntimeState{}
	}

	target := float32(0)
	if cfg.Open {
		target = 1
	}

	if !rt.Initialized {
		rt.AnimFrac = target
		rt.PrevOpen = cfg.Open
		rt.Initialized = true
		sm.Set(cfg.ID, rt)
		return cfg.Width * target
	}

	if cfg.Open != rt.PrevOpen {
		rt.PrevOpen = cfg.Open
		sm.Set(cfg.ID, rt)
		sidebarStartAnimation(cfg.ID, rt.AnimFrac, target,
			cfg.Spring, cfg.TweenDuration, cfg.TweenEasing, w)
	}

	return cfg.Width * f32Max(0, rt.AnimFrac)
}

func sidebarStartAnimation(
	sidebarID string, from, to float32,
	springCfg SpringCfg,
	tweenDur time.Duration, tweenEasing EasingFn,
	w *Window,
) {
	animID := "sidebar:" + sidebarID
	onValue := func(v float32, w *Window) {
		sm := StateMap[string, SidebarRuntimeState](
			w, nsSidebar, capFew)
		rt, _ := sm.Get(sidebarID)
		rt.AnimFrac = v
		sm.Set(sidebarID, rt)
	}
	if tweenDur > 0 {
		w.AnimationAdd(&TweenAnimation{
			AnimID:   animID,
			From:     from,
			To:       to,
			Duration: tweenDur,
			Easing:   tweenEasing,
			OnValue:  onValue,
		})
	} else {
		sp := &SpringAnimation{
			AnimID:  animID,
			Config:  springCfg,
			OnValue: onValue,
		}
		sp.SpringTo(from, to)
		w.AnimationAdd(sp)
	}
}
