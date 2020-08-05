package state

import (
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
