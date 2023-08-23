package alert

type Notifier interface {
	Notify(Alert) error
}

type PriorityNotifier struct {
	priority Notifier
	normal   Notifier
}

func (pn *PriorityNotifier) Notify(a Alert) error {
	var e error

	if a.Priority() {
		if err := pn.priority.Notify(a); err != nil {
			e = err
		}
	}

	if err := pn.normal.Notify(a); err != nil {
		e = err
	}

	return e
}

func NewPriorityNotifier(priority, normal Notifier) *PriorityNotifier {
	return &PriorityNotifier{
		priority: priority,
		normal:   normal}
}

var _ Notifier = &PriorityNotifier{}
