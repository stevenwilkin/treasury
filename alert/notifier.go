package alert

type Notifier interface {
	Notify(Alert) error
}
