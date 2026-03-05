package gui

import "fmt"

const splitterDefaultRatio = float32(0.5)

// SplitterOrientation controls how panes are arranged.
type SplitterOrientation uint8

const (
	SplitterHorizontal SplitterOrientation = iota
	SplitterVertical
)

// SplitterCollapsed tracks which pane is collapsed, if any.
type SplitterCollapsed uint8

const (
	SplitterCollapseNone   SplitterCollapsed = iota
	SplitterCollapseFirst
	SplitterCollapseSecond
)

// SplitterState is an app-owned persistence model.
type SplitterState struct {
	Ratio     float32
	Collapsed SplitterCollapsed
}

// SplitterStateNormalize normalizes state before persisting.
func SplitterStateNormalize(state SplitterState) SplitterState {
	return SplitterState{
		Ratio:     splitterNormalizeRatio(state.Ratio),
		Collapsed: state.Collapsed,
	}
}

// SplitterPaneCfg configures one pane of a splitter.
type SplitterPaneCfg struct {
	MinSize       float32
	MaxSize       float32
	Collapsible   bool
	CollapsedSize float32
	Content       []View
}

// splitterPaneCore holds pane fields needed by callbacks
// (excludes Content to avoid GC false retention).
type splitterPaneCore struct {
	minSize       float32
	maxSize       float32
	collapsible   bool
	collapsedSize float32
}

// SplitterCfg configures a splitter component.
type SplitterCfg struct {
	ID                    string
	IDFocus               uint32
	Orientation           SplitterOrientation
	Sizing                Sizing
	Ratio                 float32
	Collapsed             SplitterCollapsed
	OnChange              func(float32, SplitterCollapsed, *Event, *Window)
	First                 SplitterPaneCfg
	Second                SplitterPaneCfg
	HandleSize            float32
	DragStep              float32
	DragStepLarge         float32
	DoubleClickCollapse   bool
	ShowCollapseButtons   bool
	ColorHandle           Color
	ColorHandleHover      Color
	ColorHandleActive     Color
	ColorHandleBorder     Color
	ColorGrip             Color
	ColorButton           Color
	ColorButtonHover      Color
	ColorButtonActive     Color
	ColorButtonIcon       Color
	SizeBorder            float32
	Radius                float32
	RadiusBorder          float32
	Disabled              bool
	Invisible             bool

	A11YLabel       string
	A11YDescription string
}

// splitterCore holds callback-relevant fields.
type splitterCore struct {
	id                    string
	idFocus               uint32
	orientation           SplitterOrientation
	ratio                 float32
	collapsed             SplitterCollapsed
	onChange              func(float32, SplitterCollapsed, *Event, *Window)
	first                 splitterPaneCore
	second                splitterPaneCore
	handleSize            float32
	dragStep              float32
	dragStepLarge         float32
	doubleClickCollapse   bool
	disabled              bool
}

type splitterComputed struct {
	firstMain  float32
	secondMain float32
	handleMain float32
	ratio      float32
	collapsed  SplitterCollapsed
}

func newSplitterCore(cfg *SplitterCfg) *splitterCore {
	return &splitterCore{
		id:          cfg.ID,
		idFocus:     cfg.IDFocus,
		orientation: cfg.Orientation,
		ratio:       cfg.Ratio,
		collapsed:   cfg.Collapsed,
		onChange:     cfg.OnChange,
		first: splitterPaneCore{
			minSize:       cfg.First.MinSize,
			maxSize:       cfg.First.MaxSize,
			collapsible:   cfg.First.Collapsible,
			collapsedSize: cfg.First.CollapsedSize,
		},
		second: splitterPaneCore{
			minSize:       cfg.Second.MinSize,
			maxSize:       cfg.Second.MaxSize,
			collapsible:   cfg.Second.Collapsible,
			collapsedSize: cfg.Second.CollapsedSize,
		},
		handleSize:          cfg.HandleSize,
		dragStep:            cfg.DragStep,
		dragStepLarge:       cfg.DragStepLarge,
		doubleClickCollapse: cfg.DoubleClickCollapse,
		disabled:            cfg.Disabled,
	}
}

func applySplitterDefaults(cfg *SplitterCfg) {
	s := &DefaultSplitterStyle
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FillFill
	}
	if cfg.Ratio == 0 {
		cfg.Ratio = splitterDefaultRatio
	}
	if cfg.HandleSize == 0 {
		cfg.HandleSize = s.HandleSize
	}
	if cfg.DragStep == 0 {
		cfg.DragStep = s.DragStep
	}
	if cfg.DragStepLarge == 0 {
		cfg.DragStepLarge = s.DragStepLarge
	}
	if !cfg.ColorHandle.IsSet() {
		cfg.ColorHandle = s.ColorHandle
	}
	if !cfg.ColorHandleHover.IsSet() {
		cfg.ColorHandleHover = s.ColorHandleHover
	}
	if !cfg.ColorHandleActive.IsSet() {
		cfg.ColorHandleActive = s.ColorHandleActive
	}
	if !cfg.ColorHandleBorder.IsSet() {
		cfg.ColorHandleBorder = s.ColorHandleBorder
	}
	if !cfg.ColorGrip.IsSet() {
		cfg.ColorGrip = s.ColorGrip
	}
	if !cfg.ColorButton.IsSet() {
		cfg.ColorButton = s.ColorButton
	}
	if !cfg.ColorButtonHover.IsSet() {
		cfg.ColorButtonHover = s.ColorButtonHover
	}
	if !cfg.ColorButtonActive.IsSet() {
		cfg.ColorButtonActive = s.ColorButtonActive
	}
	if !cfg.ColorButtonIcon.IsSet() {
		cfg.ColorButtonIcon = s.ColorButtonIcon
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = s.SizeBorder
	}
	if cfg.Radius == 0 {
		cfg.Radius = s.Radius
	}
	if cfg.RadiusBorder == 0 {
		cfg.RadiusBorder = s.RadiusBorder
	}
}

// Split is an alias for Splitter.
func Split(cfg SplitterCfg) View {
	return Splitter(cfg)
}

// Splitter creates a two-pane splitter with drag/keyboard/collapse.
func Splitter(cfg SplitterCfg) View {
	applySplitterDefaults(&cfg)
	core := newSplitterCore(&cfg)

	return Canvas(ContainerCfg{
		ID:              cfg.ID,
		IDFocus:         cfg.IDFocus,
		A11YRole:        AccessRoleSplitter,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.ID),
		A11YDescription: cfg.A11YDescription,
		Sizing:          cfg.Sizing,
		Padding:         PaddingNone,
		Clip:            true,
		Disabled:        cfg.Disabled,
		Invisible:       cfg.Invisible,
		OnKeyDown: func(_ *Layout, e *Event, w *Window) {
			splitterOnKeydown(core, e, w)
		},
		AmendLayout: func(layout *Layout, w *Window) {
			splitterAmendLayout(core, layout, w)
		},
		Content: []View{
			splitterPane(fmt.Sprintf("%s:pane:first", cfg.ID),
				cfg.First.Content),
			splitterHandleView(&cfg, core),
			splitterPane(fmt.Sprintf("%s:pane:second", cfg.ID),
				cfg.Second.Content),
		},
	})
}

func splitterPane(id string, content []View) View {
	return Column(ContainerCfg{
		ID:      id,
		Sizing:  FixedFixed,
		Padding: PaddingNone,
		Clip:    true,
		Content: content,
	})
}

func splitterHandleView(cfg *SplitterCfg, core *splitterCore) View {
	content := make([]View, 0, 3)
	if cfg.ShowCollapseButtons &&
		(cfg.First.Collapsible || cfg.Second.Collapsible) {
		if cfg.First.Collapsible {
			content = append(content,
				splitterButton(cfg, core, SplitterCollapseFirst))
		}
		content = append(content, splitterGrip(cfg))
		if cfg.Second.Collapsible {
			content = append(content,
				splitterButton(cfg, core, SplitterCollapseSecond))
		}
	} else {
		content = append(content, splitterGrip(cfg))
	}

	orientation := cfg.Orientation
	colorHover := cfg.ColorHandleHover
	colorActive := cfg.ColorHandleActive

	var handleWidth, handleHeight float32
	if orientation == SplitterHorizontal {
		handleWidth = cfg.HandleSize
	} else {
		handleHeight = cfg.HandleSize
	}

	handleCfg := ContainerCfg{
		ID:          fmt.Sprintf("%s:handle", cfg.ID),
		Sizing:      FixedFixed,
		Width:       handleWidth,
		Height:      handleHeight,
		Padding:     PaddingNone,
		Spacing:     Some[float32](1),
		Color:       cfg.ColorHandle,
		ColorBorder: cfg.ColorHandleBorder,
		SizeBorder:  Some(cfg.SizeBorder),
		Radius:      Some(cfg.Radius),
		HAlign:      HAlignCenter,
		VAlign:      VAlignMiddle,
		OnClick: func(_ *Layout, e *Event, w *Window) {
			splitterOnHandleClick(core, e, w)
		},
		OnHover: func(layout *Layout, e *Event, w *Window) {
			splitterOnHandleHover(orientation, colorHover,
				colorActive, layout, e, w)
		},
		Content: content,
	}

	if orientation == SplitterHorizontal {
		return Column(handleCfg)
	}
	return Row(handleCfg)
}

func splitterGrip(cfg *SplitterCfg) View {
	isHoriz := cfg.Orientation == SplitterHorizontal
	var w, h float32
	if isHoriz {
		w = f32Max(2, cfg.HandleSize*0.35)
		h = f32Max(14, cfg.HandleSize*2.0)
	} else {
		w = f32Max(14, cfg.HandleSize*2.0)
		h = f32Max(2, cfg.HandleSize*0.35)
	}
	return Rectangle(RectangleCfg{
		Width:  w,
		Height: h,
		Color:  cfg.ColorGrip,
		Radius: cfg.RadiusBorder,
		Sizing: FixedFixed,
	})
}

func splitterButton(cfg *SplitterCfg, core *splitterCore,
	target SplitterCollapsed) View {
	size := f32Max(4, cfg.HandleSize-2)
	ts := TextStyle{
		Color: cfg.ColorButtonIcon,
		Size:  size,
	}
	return Button(ButtonCfg{
		ID:          fmt.Sprintf("%s:button:%d", cfg.ID, target),
		Width:       size,
		Height:      size,
		Sizing:      FixedFixed,
		Padding:     PaddingNone,
		Color:       cfg.ColorButton,
		ColorHover:  cfg.ColorButtonHover,
		ColorClick:  cfg.ColorButtonActive,
		ColorFocus:  cfg.ColorButtonHover,
		Radius:      Some(cfg.RadiusBorder),
		OnClick: func(_ *Layout, e *Event, w *Window) {
			splitterOnButtonClick(core, target, e, w)
		},
		Content: []View{
			Text(TextCfg{
				Text:      splitterButtonIcon(core, target),
				TextStyle: ts,
			}),
		},
	})
}

func splitterButtonIcon(core *splitterCore, target SplitterCollapsed) string {
	current := splitterEffectiveCollapsed(core, core.collapsed)
	if core.orientation == SplitterHorizontal {
		if target == SplitterCollapseFirst {
			if current == SplitterCollapseFirst {
				return "▶"
			}
			return "◀"
		}
		if current == SplitterCollapseSecond {
			return "◀"
		}
		return "▶"
	}
	if target == SplitterCollapseFirst {
		if current == SplitterCollapseFirst {
			return "▼"
		}
		return "▲"
	}
	if current == SplitterCollapseSecond {
		return "▲"
	}
	return "▼"
}

// --- Event handlers ---

func splitterOnKeydown(core *splitterCore, e *Event, w *Window) {
	if core.disabled {
		return
	}
	ly, ok := w.layout.FindByID(core.id)
	if !ok {
		return
	}
	mainSz := splitterMainSize(ly, core.orientation)
	handle := splitterHandleSizeFromLayout(ly, core.orientation,
		core.handleSize)
	available := f32Max(0, mainSz-handle)

	nextRatio := splitterClampRatio(core, available, core.ratio)
	nextCollapsed := splitterEffectiveCollapsed(core, core.collapsed)
	handled := false

	isShift := e.Modifiers == ModShift
	isNone := e.Modifiers == ModNone

	switch e.KeyCode {
	case KeyLeft:
		if core.orientation == SplitterHorizontal && (isNone || isShift) {
			nextCollapsed = SplitterCollapseNone
			step := core.dragStep
			if isShift {
				step = core.dragStepLarge
			}
			nextRatio = splitterClampRatio(core, available,
				nextRatio-splitterStep(step))
			handled = true
		}
	case KeyRight:
		if core.orientation == SplitterHorizontal && (isNone || isShift) {
			nextCollapsed = SplitterCollapseNone
			step := core.dragStep
			if isShift {
				step = core.dragStepLarge
			}
			nextRatio = splitterClampRatio(core, available,
				nextRatio+splitterStep(step))
			handled = true
		}
	case KeyUp:
		if core.orientation == SplitterVertical && (isNone || isShift) {
			nextCollapsed = SplitterCollapseNone
			step := core.dragStep
			if isShift {
				step = core.dragStepLarge
			}
			nextRatio = splitterClampRatio(core, available,
				nextRatio-splitterStep(step))
			handled = true
		}
	case KeyDown:
		if core.orientation == SplitterVertical && (isNone || isShift) {
			nextCollapsed = SplitterCollapseNone
			step := core.dragStep
			if isShift {
				step = core.dragStepLarge
			}
			nextRatio = splitterClampRatio(core, available,
				nextRatio+splitterStep(step))
			handled = true
		}
	case KeyHome:
		if isNone && core.first.collapsible {
			nextCollapsed = SplitterCollapseFirst
			handled = true
		}
	case KeyEnd:
		if isNone && core.second.collapsible {
			nextCollapsed = SplitterCollapseSecond
			handled = true
		}
	case KeyEnter:
		if isNone {
			target := splitterToggleTarget(core, nextCollapsed)
			if target != SplitterCollapseNone {
				if nextCollapsed == target {
					nextCollapsed = SplitterCollapseNone
				} else {
					nextCollapsed = target
				}
				handled = true
			}
		}
	default:
		if e.CharCode == CharSpace && isNone {
			target := splitterToggleTarget(core, nextCollapsed)
			if target != SplitterCollapseNone {
				if nextCollapsed == target {
					nextCollapsed = SplitterCollapseNone
				} else {
					nextCollapsed = target
				}
				handled = true
			}
		}
	}

	if handled {
		splitterEmitChange(core, nextRatio, nextCollapsed, e, w)
	}
}

func splitterOnHandleClick(core *splitterCore, e *Event, w *Window) {
	if core.disabled {
		return
	}
	splitterSetCursor(core.orientation, w)
	splitterFocus(core, w)

	idFocus := core.idFocus
	w.MouseLock(MouseLockCfg{
		MouseMove: func(_ *Layout, e *Event, w *Window) {
			splitterOnDragMove(core, e, w)
		},
		MouseUp: func(_ *Layout, _ *Event, w *Window) {
			w.MouseUnlock()
			if idFocus > 0 {
				w.SetIDFocus(idFocus)
			}
		},
	})
	e.IsHandled = true
}

func splitterOnDragMove(core *splitterCore, e *Event, w *Window) {
	if core.disabled {
		return
	}
	ly, ok := w.layout.FindByID(core.id)
	if !ok {
		return
	}
	mainSz := splitterMainSize(ly, core.orientation)
	handle := splitterHandleSizeFromLayout(ly, core.orientation,
		core.handleSize)
	available := f32Max(0, mainSz-handle)
	if available <= 0 {
		return
	}

	var cursorMain float32
	if core.orientation == SplitterHorizontal {
		cursorMain = e.MouseX - ly.Shape.X - (handle / 2)
	} else {
		cursorMain = e.MouseY - ly.Shape.Y - (handle / 2)
	}
	ratio := splitterClampRatio(core, available, cursorMain/available)
	splitterSetCursor(core.orientation, w)
	splitterEmitChange(core, ratio, SplitterCollapseNone, e, w)
}

func splitterOnHandleHover(
	orientation SplitterOrientation,
	colorHover, colorActive Color,
	layout *Layout, e *Event, w *Window,
) {
	splitterSetCursor(orientation, w)
	layout.Shape.Color = colorHover
	if e.MouseButton == MouseLeft {
		layout.Shape.Color = colorActive
	}
	e.IsHandled = true
}

func splitterOnButtonClick(
	core *splitterCore,
	target SplitterCollapsed,
	e *Event, w *Window,
) {
	if core.disabled {
		return
	}
	validTarget := splitterEffectiveCollapsed(core, target)
	if validTarget == SplitterCollapseNone {
		return
	}
	ratio := splitterCurrentRatio(core, w)
	current := splitterEffectiveCollapsed(core, core.collapsed)
	next := validTarget
	if current == validTarget {
		next = SplitterCollapseNone
	}
	splitterEmitChange(core, ratio, next, e, w)
}

// --- AmendLayout ---

func splitterAmendLayout(core *splitterCore, layout *Layout, w *Window) {
	if len(layout.Children) < 3 {
		return
	}

	mainSz := splitterMainSize(layout, core.orientation)
	computed := splitterCompute(core, mainSz)

	if core.orientation == SplitterHorizontal {
		x := layout.Shape.X
		y := layout.Shape.Y
		h := layout.Shape.Height
		splitterLayoutChild(&layout.Children[0], x, y,
			computed.firstMain, h, w)
		splitterLayoutChild(&layout.Children[1],
			x+computed.firstMain, y, computed.handleMain, h, w)
		splitterLayoutChild(&layout.Children[2],
			x+computed.firstMain+computed.handleMain, y,
			computed.secondMain, h, w)
	} else {
		x := layout.Shape.X
		y := layout.Shape.Y
		wid := layout.Shape.Width
		splitterLayoutChild(&layout.Children[0], x, y,
			wid, computed.firstMain, w)
		splitterLayoutChild(&layout.Children[1], x,
			y+computed.firstMain, wid, computed.handleMain, w)
		splitterLayoutChild(&layout.Children[2], x,
			y+computed.firstMain+computed.handleMain,
			wid, computed.secondMain, w)
	}
}

func splitterLayoutChild(
	child *Layout,
	x, y, width, height float32,
	w *Window,
) {
	splitterResetPositions(child, true, AxisNone, 0, 0)
	child.Shape.Sizing = FixedFixed
	child.Shape.Width = f32Max(0, width)
	child.Shape.Height = f32Max(0, height)
	child.Shape.MinWidth = child.Shape.Width
	child.Shape.MaxWidth = child.Shape.Width
	child.Shape.MinHeight = child.Shape.Height
	child.Shape.MaxHeight = child.Shape.Height
	child.Shape.X = 0
	child.Shape.Y = 0

	layoutWidths(child)
	layoutFillWidths(child)
	layoutWrapText(child, w)
	layoutHeights(child)
	layoutFillHeights(child)
	layoutAdjustScrollOffsets(child, w)
	layoutPositions(child, x, y, w)
	layoutAmend(child, w)
}

func splitterResetPositions(layout *Layout, isRoot bool,
	parentAxis Axis, parentOldX, parentOldY float32) {
	oldX := layout.Shape.X
	oldY := layout.Shape.Y
	if isRoot {
		layout.Shape.X = 0
		layout.Shape.Y = 0
	} else if parentAxis == AxisNone {
		layout.Shape.X = oldX - parentOldX
		layout.Shape.Y = oldY - parentOldY
	} else {
		layout.Shape.X = 0
		layout.Shape.Y = 0
	}
	for i := range layout.Children {
		splitterResetPositions(&layout.Children[i], false,
			layout.Shape.Axis, oldX, oldY)
	}
}

// --- Pure computation helpers ---

func splitterCompute(core *splitterCore, mainSize float32) splitterComputed {
	handle := splitterHandleSize(core.handleSize, mainSize)
	available := f32Max(0, mainSize-handle)
	ratio := splitterClampRatio(core, available, core.ratio)
	collapsed := splitterEffectiveCollapsed(core, core.collapsed)

	var first, second float32
	switch collapsed {
	case SplitterCollapseFirst:
		first, second = splitterCollapsedFirst(core, available)
	case SplitterCollapseSecond:
		first, second = splitterCollapsedSecond(core, available)
	default:
		first = splitterClampFirstSize(core, available, ratio*available)
		second = f32Max(0, available-first)
		if available > 0 {
			ratio = first / available
		} else {
			ratio = splitterDefaultRatio
		}
	}
	return splitterComputed{
		firstMain:  first,
		secondMain: second,
		handleMain: handle,
		ratio:      ratio,
		collapsed:  collapsed,
	}
}

func splitterCollapsedFirst(core *splitterCore, available float32) (float32, float32) {
	firstTarget := f32Clamp(core.first.collapsedSize, 0, available)
	secondMin := f32Max(0, core.second.minSize)
	secondMax := splitterLimitMax(core.second.maxSize, available)
	if secondMin > secondMax {
		secondMin = secondMax
	}
	second := f32Clamp(available-firstTarget, secondMin, secondMax)
	first := f32Max(0, available-second)
	first = f32Min(first, splitterLimitMax(core.first.maxSize, available))
	second = f32Max(0, available-first)
	return first, second
}

func splitterCollapsedSecond(core *splitterCore, available float32) (float32, float32) {
	secondTarget := f32Clamp(core.second.collapsedSize, 0, available)
	firstMin := f32Max(0, core.first.minSize)
	firstMax := splitterLimitMax(core.first.maxSize, available)
	if firstMin > firstMax {
		firstMin = firstMax
	}
	first := f32Clamp(available-secondTarget, firstMin, firstMax)
	second := f32Max(0, available-first)
	second = f32Min(second, splitterLimitMax(core.second.maxSize, available))
	return f32Max(0, available-second), f32Max(0, second)
}

func splitterMainSize(layout *Layout, orientation SplitterOrientation) float32 {
	if orientation == SplitterHorizontal {
		return layout.Shape.Width
	}
	return layout.Shape.Height
}

func splitterHandleSizeFromLayout(
	layout *Layout,
	orientation SplitterOrientation,
	fallback float32,
) float32 {
	if len(layout.Children) > 1 {
		handle := layout.Children[1]
		if orientation == SplitterHorizontal {
			return handle.Shape.Width
		}
		return handle.Shape.Height
	}
	return fallback
}

func splitterHandleSize(handleSize, mainSize float32) float32 {
	size := f32Max(1, handleSize)
	if mainSize <= 0 {
		return size
	}
	return f32Min(size, mainSize)
}

func splitterClampRatio(core *splitterCore, available, ratio float32) float32 {
	if available <= 0 {
		return splitterDefaultRatio
	}
	target := splitterNormalizeRatio(ratio) * available
	first := splitterClampFirstSize(core, available, target)
	return first / available
}

func splitterClampFirstSize(core *splitterCore, available, target float32) float32 {
	lower, upper := splitterBounds(core, available)
	lower = f32Clamp(lower, 0, available)
	upper = f32Clamp(upper, 0, available)
	if lower <= upper {
		return f32Clamp(target, lower, upper)
	}
	return f32Clamp(target, upper, lower)
}

func splitterBounds(core *splitterCore, available float32) (float32, float32) {
	firstMin := f32Max(0, core.first.minSize)
	firstMax := splitterLimitMax(core.first.maxSize, available)
	if firstMin > firstMax {
		firstMin = firstMax
	}
	secondMin := f32Max(0, core.second.minSize)
	secondMax := splitterLimitMax(core.second.maxSize, available)
	if secondMin > secondMax {
		secondMin = secondMax
	}
	lower := f32Max(firstMin, available-secondMax)
	upper := f32Min(firstMax, available-secondMin)
	return lower, upper
}

func splitterLimitMax(value, available float32) float32 {
	if value > 0 {
		return f32Clamp(value, 0, available)
	}
	return available
}

func splitterNormalizeRatio(ratio float32) float32 {
	return f32Clamp(ratio, 0, 1)
}

func splitterCurrentRatio(core *splitterCore, w *Window) float32 {
	ly, ok := w.layout.FindByID(core.id)
	if !ok {
		return splitterNormalizeRatio(core.ratio)
	}
	mainSz := splitterMainSize(ly, core.orientation)
	handle := splitterHandleSizeFromLayout(ly, core.orientation,
		core.handleSize)
	return splitterClampRatio(core, f32Max(0, mainSz-handle), core.ratio)
}

func splitterToggleTarget(core *splitterCore, current SplitterCollapsed) SplitterCollapsed {
	active := splitterEffectiveCollapsed(core, current)
	if active != SplitterCollapseNone {
		return active
	}
	if core.first.collapsible {
		return SplitterCollapseFirst
	}
	if core.second.collapsible {
		return SplitterCollapseSecond
	}
	return SplitterCollapseNone
}

func splitterStep(step float32) float32 {
	if step > 0 {
		return step
	}
	return 0.02
}

func splitterEffectiveCollapsed(core *splitterCore, collapsed SplitterCollapsed) SplitterCollapsed {
	switch collapsed {
	case SplitterCollapseFirst:
		if core.first.collapsible {
			return SplitterCollapseFirst
		}
		return SplitterCollapseNone
	case SplitterCollapseSecond:
		if core.second.collapsible {
			return SplitterCollapseSecond
		}
		return SplitterCollapseNone
	default:
		return SplitterCollapseNone
	}
}

func splitterEmitChange(
	core *splitterCore,
	ratio float32, collapsed SplitterCollapsed,
	e *Event, w *Window,
) {
	state := SplitterStateNormalize(SplitterState{
		Ratio:     ratio,
		Collapsed: collapsed,
	})
	if core.onChange != nil {
		core.onChange(state.Ratio, state.Collapsed, e, w)
	}
	splitterFocus(core, w)
	e.IsHandled = true
}

func splitterFocus(core *splitterCore, w *Window) {
	if core.idFocus > 0 {
		w.SetIDFocus(core.idFocus)
	}
}

func splitterSetCursor(orientation SplitterOrientation, w *Window) {
	if orientation == SplitterHorizontal {
		w.SetMouseCursorEW()
	} else {
		w.SetMouseCursorNS()
	}
}
