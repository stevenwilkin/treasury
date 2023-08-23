package alert

import (
	"errors"
	"testing"
)

type FailingTestNotifier struct{}

func (n *FailingTestNotifier) Notify(_ Alert) error { return errors.New("Fail") }

var _ Notifier = &FailingTestNotifier{}

func TestPriorityNotifierWithNormalAlert(t *testing.T) {
	normalAlert := &TestAlert{}

	priority := &TestNotifier{}
	normal := &TestNotifier{}

	notifier := NewPriorityNotifier(priority, normal)
	notifier.Notify(normalAlert)

	if priority.alert != nil {
		t.Error("Should not send priority notification")
	}

	if normal.alert != normalAlert {
		t.Error("Should send normal notification")
	}
}

func TestPriorityNotifierWithPriorityAlert(t *testing.T) {
	priorityAlert := &TestAlert{priority: true}

	priority := &TestNotifier{}
	normal := &TestNotifier{}

	notifier := NewPriorityNotifier(priority, normal)
	notifier.Notify(priorityAlert)

	if priority.alert != priorityAlert {
		t.Error("Should send priority notification")
	}

	if normal.alert != priorityAlert {
		t.Error("Should send normal notification")
	}
}

func TestPriorityNotifierWithFailingPriorityNotification(t *testing.T) {
	priorityAlert := &TestAlert{priority: true}

	priority := &FailingTestNotifier{}
	normal := &TestNotifier{}

	notifier := NewPriorityNotifier(priority, normal)
	err := notifier.Notify(priorityAlert)

	if err == nil {
		t.Error("Should return an error")
	}

	if normal.alert != priorityAlert {
		t.Error("Should send normal notification")
	}
}
