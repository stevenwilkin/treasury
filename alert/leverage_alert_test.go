package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
)

func TestLeverageAlertDescription(t *testing.T) {
	alert := NewLeverageAlert(nil, 4)

	expected := "Leverage alert at 4.00"
	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}

func TestLeverageAlertMessage(t *testing.T) {
	state := state.NewState()
	alert := NewLeverageAlert(state, 4)

	state.SetLeverageDeribit(2)
	state.SetLeverageBybit(3)

	expected := "Leverage: 2.00 3.00"
	if alert.Message() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Message())
	}
}

func TestLeverageAlertDeactivate(t *testing.T) {
	alert := NewLeverageAlert(nil, 4)

	if !alert.Active() {
		t.Error("Alert should be active")
	}

	alert.Deactivate()

	if alert.Active() {
		t.Error("Alert should be inactive")
	}
}

func TestLeverageAlertCheck(t *testing.T) {
	state := state.NewState()
	alert := NewLeverageAlert(state, 4)

	state.SetLeverageDeribit(3)
	state.SetLeverageBybit(3.9)

	if alert.Check() {
		t.Error("Alert should not be triggered")
	}

	state.SetLeverageDeribit(4)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}
