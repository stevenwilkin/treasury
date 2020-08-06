package state

import (
	"sync"
	"testing"

	"github.com/stevenwilkin/treasury/symbol"
)

func TestSubscribeToSymbols(t *testing.T) {
	s := NewState()
	ch := s.SubscribeToSymbols()

	var notification SymbolNotification

	go func() {
		notification = <-ch
	}()

	s.SetSymbol(symbol.BTCTHB, 300000)

	if notification.Symbol != symbol.BTCTHB || notification.Value != 300000 {
		t.Error("Expected channel to have been written to")
	}
}

func TestSubscribeToSymbolsDuplicateEvents(t *testing.T) {
	s := NewState()
	ch := s.SubscribeToSymbols()

	wg := sync.WaitGroup{}
	wg.Add(1)
	count := 0

	go func() {
		for {
			if (<-ch == SymbolNotification{}) {
				break
			}
			count++
		}
		wg.Done()
	}()

	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetSymbol(symbol.BTCTHB, 300000)
	close(ch)
	wg.Wait()

	if count != 1 {
		t.Errorf("Expected 1 write to channel, got %d", count)
	}
}

func TestSubscribeToSymbolsMultipleEventsWithDuplicates(t *testing.T) {
	s := NewState()
	ch := s.SubscribeToSymbols()

	wg := sync.WaitGroup{}
	wg.Add(1)
	count := 0

	go func() {
		for {
			if (<-ch == SymbolNotification{}) {
				break
			}
			count++
		}
		wg.Done()
	}()

	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetSymbol(symbol.BTCTHB, 300001)
	close(ch)
	wg.Wait()

	if count != 2 {
		t.Errorf("Expected 2 writes to channel, got %d", count)
	}
}
