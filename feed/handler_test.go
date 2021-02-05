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

	if h.Status()[BTCUSDT].LastUpdate != (time.Time{}) {
		t.Error("Should not have a last update")
	}

	trigger <- true

	if h.Status()[BTCUSDT].LastUpdate == (time.Time{}) {
		t.Error("Should have a last update")
	}
}

func TestUpdateClearsErrorCountAndSetsActive(t *testing.T) {
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
	h.setFailed(BTCUSDT)

	trigger <- true

	if h.Status()[BTCUSDT].Errors != 0 {
		t.Error("Should have 0 errors")
	}

	if !h.Status()[BTCUSDT].Active {
		t.Error("Should be active")
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

	if !h.Status()[BTCUSDT].Active {
		t.Error("Should be active")
	}

	trigger <- true

	if h.Status()[BTCUSDT].Active {
		t.Error("Should not be active")
	}

	if h.Status()[BTCUSDT].Errors != 1 {
		t.Error("Should have 1 error")
	}
}

func TestClosingChannelRestart(t *testing.T) {
	delayBase = 0

	count := 0
	f := func() chan int {
		count += 1
		ch := make(chan int)
		close(ch)
		return ch
	}

	h := NewHandler()
	h.Add(BTCUSDT, f, func(int) {})

	time.Sleep(time.Millisecond) // nasty

	if count < 2 {
		t.Error("Feed should have been restarted after failure")
	}

	if count > maxRetries+1 {
		t.Error("Feed should not exceed max retries")
	}
}
