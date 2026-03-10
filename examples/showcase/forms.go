package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mike-ward/go-gui/gui"
)

type FormModel struct {
	Username string
	Email    string
	AgeText  string
	AgeValue gui.Opt[float64]

	UsernameState ShowcaseFieldState
	EmailState    ShowcaseFieldState
	AgeState      ShowcaseFieldState

	SubmitMessage string
	validationSeq int
}

type ShowcaseFieldState struct {
	Touched bool
	Dirty   bool
	Pending bool
	Issue   string
}

type ShowcaseFormSummary struct {
	Valid        bool
	Pending      bool
	InvalidCount int
	PendingCount int
}

func formSummary(form FormModel) ShowcaseFormSummary {
	states := []ShowcaseFieldState{
		form.UsernameState,
		form.EmailState,
		form.AgeState,
	}
	summary := ShowcaseFormSummary{Valid: true}
	for _, state := range states {
		if state.Issue != "" {
			summary.InvalidCount++
			summary.Valid = false
		}
		if state.Pending {
			summary.PendingCount++
			summary.Pending = true
			summary.Valid = false
		}
	}
	return summary
}

func formPendingFields(form FormModel) []string {
	fields := make([]string, 0, 3)
	if form.UsernameState.Pending {
		fields = append(fields, "username")
	}
	if form.EmailState.Pending {
		fields = append(fields, "email")
	}
	if form.AgeState.Pending {
		fields = append(fields, "age")
	}
	return fields
}

func validateUsernameSync(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "username required"
	}
	if len(value) < 3 {
		return "username min length is 3"
	}
	return ""
}

func validateEmailSync(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "email required"
	}
	if !strings.Contains(value, "@") {
		return "email must contain @"
	}
	return ""
}

func validateAgeSync(value string) string {
	if strings.TrimSpace(value) == "" {
		return "age required"
	}
	return ""
}

func validateUsernameReserved(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "admin", "root", "system":
		return "username already taken"
	default:
		return ""
	}
}

func applyFormTouch(state ShowcaseFieldState, value, initialValue, issue string) ShowcaseFieldState {
	state.Touched = true
	state.Dirty = strings.TrimSpace(value) != strings.TrimSpace(initialValue)
	state.Issue = issue
	return state
}

func startUsernameAsyncValidation(name string, w *gui.Window) {
	app := gui.State[ShowcaseApp](w)
	app.Form.validationSeq++
	seq := app.Form.validationSeq
	app.Form.UsernameState.Pending = app.Form.UsernameState.Issue == ""

	go func(name string, seq int, w *gui.Window) {
		for range 4 {
			time.Sleep(60 * time.Millisecond)
		}

		issue := validateUsernameReserved(name)

		w.QueueCommand(func(w *gui.Window) {
			app := gui.State[ShowcaseApp](w)
			if app.Form.validationSeq != seq {
				return
			}
			if strings.TrimSpace(app.Form.Username) != strings.TrimSpace(name) {
				return
			}
			app.Form.UsernameState.Pending = false
			if app.Form.UsernameState.Issue == "" {
				app.Form.UsernameState.Issue = issue
			}
			w.UpdateWindow()
		})
	}(name, seq, w)
}

func resetShowcaseForm(app *ShowcaseApp) {
	app.Form = FormModel{
		SubmitMessage: "Form reset",
	}
}

func submitShowcaseForm(w *gui.Window) {
	app := gui.State[ShowcaseApp](w)
	form := &app.Form

	form.UsernameState = applyFormTouch(form.UsernameState, form.Username, "", validateUsernameSync(form.Username))
	form.EmailState = applyFormTouch(form.EmailState, form.Email, "", validateEmailSync(form.Email))
	form.AgeState = applyFormTouch(form.AgeState, form.AgeText, "", validateAgeSync(form.AgeText))

	summary := formSummary(*form)
	switch {
	case summary.Pending:
		form.SubmitMessage = "Validation still pending"
	case !summary.Valid:
		form.SubmitMessage = fmt.Sprintf("Validation blocked submit: invalid=%d", summary.InvalidCount)
	default:
		form.SubmitMessage = fmt.Sprintf("Submitted username=%s, email=%s", strings.TrimSpace(form.Username), strings.TrimSpace(form.Email))
	}
}
