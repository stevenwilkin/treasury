package alert

import "testing"

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

func (n *TestNotifier) Notify(s string) bool { n.message = s; return true }

var _ Notifier = &TestNotifier{}

func TestChecksActiveAlerts(t *testing.T) {
	alerter := NewAlerter(&TestNotifier{})
	alert := &TestAlert{active: true}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if !alert.checked {
		t.Error("Should check alert")
	}
}

func TestDoesNotCheckInactiveAlerts(t *testing.T) {
	alerter := NewAlerter(&TestNotifier{})
	alert := &TestAlert{active: false}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if alert.checked {
		t.Error("Should not check inactive alert")
	}
}

func TestNotifiesTriggeredAlert(t *testing.T) {
	notifier := &TestNotifier{}
	alerter := NewAlerter(notifier)
	alert := &TestAlert{active: true, triggered: true, message: "Foo"}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if notifier.message != "Foo" {
		t.Error("Should notify triggered alert")
	}
}

func TestDeactivatesTriggeredAlert(t *testing.T) {
	alerter := NewAlerter(&TestNotifier{})
	alert := &TestAlert{active: true, triggered: true}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if alert.active {
		t.Error("Should deactivate triggered alert")
	}
}
