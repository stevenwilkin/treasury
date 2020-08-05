package state

import (
	"github.com/stevenwilkin/treasury/symbol"
)

type SymbolNotification struct {
	Symbol symbol.Symbol
	Value  float64
}

func (s *State) SubscribeToSymbols() chan SymbolNotification {
	c := make(chan SymbolNotification)

	s.symbolSubscribers[c] = true

	return c
}

func (s *State) NotifySymbolSubscribers(sym symbol.Symbol, value float64) {
	for ch, _ := range s.symbolSubscribers {
		ch <- SymbolNotification{Symbol: sym, Value: value}
	}
}
