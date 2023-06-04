package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
)

func TestFundingRateAlertDeactivate(t *testing.T) {
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

func TestFundingRateAlertMessage(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFundingRate(0.001)

	expected := "Funding: 0.100000%"
	if alert.Message() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Message())
	}
}

func TestCheckOnPositiveFundingRate(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFundingRate(0.1)

	if alert.Check() {
		t.Error("Alert should not be triggered")
	}
}

func TestCheckOnNegativeFundingRate(t *testing.T) {
	state := state.NewState()
	alert := NewFundingAlert(state)

	state.SetFundingRate(-0.1)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}
