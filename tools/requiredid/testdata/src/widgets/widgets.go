package widgets

type WidgetCfg struct {
	ID   string `gui:"required"`
	Name string
}

type NoReqCfg struct {
	ID string
}

type S struct{}

func (S) Widget(_ WidgetCfg) {}

func Widget(_ WidgetCfg) {}
func helper(_ WidgetCfg) {}
func useN(_ NoReqCfg)    {}

func good() {
	Widget(WidgetCfg{ID: "ok", Name: "x"})
}

func missingID() {
	Widget(WidgetCfg{Name: "x"}) // want `WidgetCfg.ID is required`
}

func emptyID() {
	Widget(WidgetCfg{ID: "", Name: "x"}) // want `WidgetCfg.ID is required`
}

func methodCall() {
	var s S
	s.Widget(WidgetCfg{Name: "x"}) // want `WidgetCfg.ID is required`
}

func noTagIgnored() {
	useN(NoReqCfg{})
}

func ignoredByDirective() {
	Widget(WidgetCfg{Name: "x"}) //requiredid:ignore
}

func helperArgSkipped() {
	helper(WidgetCfg{Name: "x"})
}

func returnedSkipped() WidgetCfg {
	return WidgetCfg{Name: "x"}
}

func varAssignSkipped() {
	_ = WidgetCfg{Name: "x"}
}
