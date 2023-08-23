package alert

import (
	"github.com/stevenwilkin/treasury/state"

	log "github.com/sirupsen/logrus"
)

type Alert interface {
	Check() bool
	Active() bool
	Priority() bool
	Deactivate()
	Description() string
	Message() string
}

type Alerter struct {
	state    *state.State
	notifier Notifier
	alerts   map[Alert]bool
}

func (a *Alerter) Alerts() []Alert {
	alerts := make([]Alert, len(a.alerts))
	i := 0

	for alert := range a.alerts {
		alerts[i] = alert
		i++
	}

	return alerts
}

func (a *Alerter) ClearAlerts() {
	a.alerts = map[Alert]bool{}
}

func (a *Alerter) AddAlert(alert Alert) {
	a.alerts[alert] = true
}

func (a *Alerter) CheckAlerts() {
	for alert := range a.alerts {
		if !alert.Active() {
			continue
		}
		if alert.Check() {
			if err := a.notifier.Notify(alert); err != nil {
				log.Error(err.Error())
			}
			alert.Deactivate()
		}
	}
}

func (a *Alerter) Persist() {
	var (
		fundingAlert bool
		priceAlerts  []float64
	)

	for alert := range a.alerts {
		if !alert.Active() {
			continue
		}

		switch alert.(type) {
		case *FundingAlert:
			fundingAlert = true
		case *PriceAlert:
			priceAlerts = append(priceAlerts, alert.(*PriceAlert).price)
		}
	}

	a.state.SetFundingAlert(fundingAlert)
	a.state.SetPriceAlerts(priceAlerts)
}

func (a *Alerter) Retrieve() {
	if a.state.GetFundingAlert() {
		a.AddFundingAlert()
	}

	if priceAlerts := a.state.GetPriceAlerts(); len(priceAlerts) > 0 {
		for _, price := range priceAlerts {
			a.AddPriceAlert(price)
		}
	}
}

func NewAlerter(state *state.State, notifier Notifier) *Alerter {
	return &Alerter{
		state:    state,
		notifier: notifier,
		alerts:   map[Alert]bool{}}
}
