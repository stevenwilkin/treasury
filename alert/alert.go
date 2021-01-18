package alert

import (
	log "github.com/sirupsen/logrus"
)

type Alert interface {
	Check() bool
	Active() bool
	Deactivate()
	Description() string
	Message() string
}

type Notifier interface {
	Notify(string) error
}

type Alerter struct {
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
			if err := a.notifier.Notify(alert.Message()); err != nil {
				log.Error(err.Error())
			}
			alert.Deactivate()
		}
	}
}

func NewAlerter(notifier Notifier) *Alerter {
	return &Alerter{
		notifier: notifier,
		alerts:   map[Alert]bool{}}
}
