package alert

import (
	"testing"

	"github.com/stevenwilkin/treasury/state"
)

type TestAlert struct {
	active    bool
	priority  bool
	triggered bool
	checked   bool
}

func (a *TestAlert) Check() bool         { a.checked = true; return a.triggered }
func (a *TestAlert) Active() bool        { return a.active }
func (a *TestAlert) Priority() bool      { return a.priority }
func (a *TestAlert) Deactivate()         { a.active = false }
func (a *TestAlert) Description() string { return "" }
func (a *TestAlert) Message() string     { return "" }

var _ Alert = &TestAlert{}

type TestNotifier struct {
	alert Alert
}

func (n *TestNotifier) Notify(a Alert) error { n.alert = a; return nil }

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
	alert := &TestAlert{active: true, triggered: true}

	alerter.AddAlert(alert)
	alerter.CheckAlerts()

	if notifier.alert != alert {
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

func TestDoesNotPersistInactiveAlerts(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})
	alerter.AddAlert(&PriceAlert{})
	alerter.AddAlert(&FundingAlert{})

	alerter.Persist()

	if s.GetFundingAlert() {
		t.Error("Should not have funding alert")
	}

	if len(s.GetPriceAlerts()) != 0 {
		t.Error("Should not have price alerts")
	}
}

func TestPersistsFundingAlert(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})
	alerter.AddAlert(&FundingAlert{active: true})

	alerter.Persist()

	if !s.GetFundingAlert() {
		t.Error("Should have funding alert")
	}
}

func TestPersistsPriceAlerts(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})
	alerter.AddAlert(&PriceAlert{active: true, price: 10000})
	alerter.AddAlert(&PriceAlert{active: true, price: 20000})

	alerter.Persist()

	if len(s.GetPriceAlerts()) != 2 {
		t.Fatal("Should have price alerts")
	}

	var hasFirstAlert, hasSecondAlert bool

	for _, a := range s.GetPriceAlerts() {
		if a == 10000 {
			hasFirstAlert = true
		} else if a == 20000 {
			hasSecondAlert = true
		}
	}

	if !hasFirstAlert || !hasSecondAlert {
		t.Error("Should persist details of price alerts")
	}
}

func TestPersistClearsPreviousAlerts(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})

	s.SetFundingAlert(true)
	s.SetPriceAlerts([]float64{10000})

	alerter.Persist()

	if s.GetFundingAlert() {
		t.Error("Should not have funding alert")
	}

	if len(s.GetPriceAlerts()) != 0 {
		t.Error("Should not have price alerts")
	}
}

func TestRetrieve(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})

	s.SetFundingAlert(true)
	s.SetPriceAlerts([]float64{10000, 20000})

	alerter.Retrieve()
	alerts := alerter.Alerts()

	if len(alerts) != 3 {
		t.Fatal("Should have alerts")
	}
}

func TestRetrieveFundingAlert(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})

	s.SetFundingAlert(true)

	alerter.Retrieve()
	alerts := alerter.Alerts()

	if len(alerts) != 1 {
		t.Fatal("Should have alert")
	}

	if _, ok := alerts[0].(*FundingAlert); !ok {
		t.Error("Should have a funding alert")
	}
}

func TestRetrievePriceAlerts(t *testing.T) {
	s := state.NewState()
	alerter := NewAlerter(s, &TestNotifier{})

	s.SetPriceAlerts([]float64{10000})

	alerter.Retrieve()
	alerts := alerter.Alerts()

	if len(alerts) != 1 {
		t.Fatal("Should have alert")
	}

	if alert, _ := alerts[0].(*PriceAlert); alert.price != 10000 {
		t.Error("Should have a price alert")
	}
}
