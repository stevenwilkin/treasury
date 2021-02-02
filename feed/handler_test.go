package feed

import (
	"testing"
	"time"
)

func TestLastUpdate(t *testing.T) {
	trigger := make(chan bool)
	f := func() chan int {
		ch := make(chan int)
		go func() {
			<-trigger
			ch <- 1
		}()
		return ch
	}

	h := NewHandler()
	h.Add(BTCUSDT, f, func(int) {})

	if h.feeds[BTCUSDT].LastUpdate != (time.Time{}) {
		t.Error("Should not have a last update")
	}

	trigger <- true

	if h.feeds[BTCUSDT].LastUpdate == (time.Time{}) {
		t.Error("Should have a last update")
	}
}
