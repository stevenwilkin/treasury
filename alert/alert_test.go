package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
)

type TestAlert struct {
	active    bool
	triggered bool
	checked   bool
	message   string
}

func (a *TestAlert) Check() bool         { a.checked = true; return a.triggered }
func (a *TestAlert) Active() bool        { return a.active }
func (a *TestAlert) Deactivate()         { a.active = false }
func (a *TestAlert) Description() string { return "" }
func (a *TestAlert) Message() string     { return a.message }

var _ Alert = &TestAlert{}

type TestNotifier struct {
	message string
}

func (n *TestNotifier) Notify(s string) error { n.message = s; return nil }

var _ Notifier = &TestNotifier{}

func TestAlerts(t *testing.T) {
	alerter := NewAlerter(state.NewState(), &TestNotifier{})
	alert := &TestAlert{}
	alerter.AddAlert(alert)

	if len(alerter.Alerts()) != 1 {
		t.Errorf("Should return 1 alert, got %d", len(alerter.Alerts()))
	}
}

func TestClearAlerts(t *testing.T) {
	alerter := NewAlerter(state.NewState(), &TestNotifier{})
	alert := &TestAlert{}
	alerter.AddAlert(alert)
	alerter.ClearAlerts()

	if len(alerter.Alerts()) != 0 {
		t.Errorf("Should return 0 alerts, got %d", len(alerter.Alerts()))
	}
}

func TestChecksActiveAlerts(t *testing.T) {
	alerter := NewAlerter(state.NewState(), &TestNotifier{})
	alert := &TestAlert{active: true}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if !alert.checked {
		t.Error("Should check alert")
	}
}

func TestDoesNotCheckInactiveAlerts(t *testing.T) {
	alerter := NewAlerter(state.NewState(), &TestNotifier{})
	alert := &TestAlert{active: false}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if alert.checked {
		t.Error("Should not check inactive alert")
	}
}

func TestNotifiesTriggeredAlert(t *testing.T) {
	notifier := &TestNotifier{}
	alerter := NewAlerter(state.NewState(), notifier)
	alert := &TestAlert{active: true, triggered: true, message: "Foo"}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if notifier.message != "Foo" {
		t.Error("Should notify triggered alert")
	}
}

func TestDeactivatesTriggeredAlert(t *testing.T) {
	alerter := NewAlerter(state.NewState(), &TestNotifier{})
	alert := &TestAlert{active: true, triggered: true}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if alert.active {
		t.Error("Should deactivate triggered alert")
	}
}
