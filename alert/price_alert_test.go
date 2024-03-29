package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"
)

func TestDescription(t *testing.T) {
	state := state.NewState()
	alert := NewPriceAlert(state, symbol.BTCTHB, 300000)

	expected := "Price alert at BTCTHB 300000.00"

	if alert.Description() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Description())
	}
}

func TestMessage(t *testing.T) {
	state := state.NewState()
	alert := NewPriceAlert(state, symbol.BTCTHB, 300000)

	expected := "BTCTHB has reached 300000.00"

	if alert.Message() != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, alert.Message())
	}
}

func TestDeactivate(t *testing.T) {
	state := state.NewState()
	alert := NewPriceAlert(state, symbol.BTCTHB, 300000)

	if !alert.Active() {
		t.Error("Alert should be active")
	}

	alert.Deactivate()

	if alert.Active() {
		t.Error("Alert should be inactive")
	}
}

func TestCheckOnRisingPrice(t *testing.T) {
	state := state.NewState()
	state.SetSymbol(symbol.BTCTHB, 200000)
	alert := NewPriceAlert(state, symbol.BTCTHB, 300000)

	if alert.Check() {
		t.Error("Alert should not be triggered")
	}

	state.SetSymbol(symbol.BTCTHB, 400000)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}

func TestCheckOnFallingPrice(t *testing.T) {
	state := state.NewState()
	state.SetSymbol(symbol.BTCTHB, 400000)
	alert := NewPriceAlert(state, symbol.BTCTHB, 300000)

	if alert.Check() {
		t.Error("Alert should not be triggered")
	}

	state.SetSymbol(symbol.BTCTHB, 200000)

	if !alert.Check() {
		t.Error("Alert should be triggered")
	}
}
