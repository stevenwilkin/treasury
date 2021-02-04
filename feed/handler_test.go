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

func TestUpdateClearsErrorCount(t *testing.T) {
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
	h.feeds[BTCUSDT].Errors = 1

	trigger <- true

	if h.feeds[BTCUSDT].Errors != 0 {
		t.Error("Should have 0 errors ")
	}
}

func TestClosingChannel(t *testing.T) {
	trigger := make(chan bool)
	f := func() chan int {
		ch := make(chan int)
		go func() {
			<-trigger
			close(ch)
			return
		}()
		return ch
	}

	h := NewHandler()
	h.Add(BTCUSDT, f, func(int) {})

	if !h.feeds[BTCUSDT].Active {
		t.Error("Should be active")
	}

	trigger <- true

	if h.feeds[BTCUSDT].Active {
		t.Error("Should not be active")
	}
}
