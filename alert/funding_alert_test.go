package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
)

func TestFundingAlertDeactivate(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	if !alert.Active() {
		t.Error("Alert should be active")
	}

	alert.Deactivate()

	if alert.Active() {
		t.Error("Alert should be inactive")
	}
}

func TestFundingAlertMessage(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFunding(0.001, 0.001)

	expected := "Funding: 0.100000%, Predicted: 0.100000%"
	if alert.Message() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Message())
	}
}

func TestCheckOnPositiveFunding(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFunding(0.1, 0.1)

	if alert.Check() {
		t.Error("Alert should not be triggered")
	}
}

func TestCheckOnNegativeFunding(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFunding(-0.1, 0.0)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}

func TestCheckOnNegativePredictedFunding(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFunding(0.1, -0.1)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}
