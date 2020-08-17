package alert

type Alert interface {
	Check() bool
	Active() bool
	Deactivate()
	Description() string
	Message() string
}

type Notifier interface {
	Notify(string) bool
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
			a.notifier.Notify(alert.Message())
			alert.Deactivate()
		}
	}
}

func NewAlerter(notifier Notifier) *Alerter {
	return &Alerter{
		notifier: notifier,
		alerts:   map[Alert]bool{}}
}
