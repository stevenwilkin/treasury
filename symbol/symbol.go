package symbol

import (
	"errors"
	"strings"
)

type Symbol int

const (
	BTCTHB Symbol = iota
	USDTTHB
	USDTHB
	BTCUSDT
)

type Prices map[Symbol]float64

func symbols() []string {
	return []string{"BTCTHB", "USDTTHB", "USDTHB", "BTCUSDT"}
}

func (s Symbol) String() string {
	return symbols()[s]
}

func FromString(s string) (Symbol, error) {
	for i, symbol := range symbols() {
		if strings.ToLower(s) == strings.ToLower(symbol) {
			return Symbol(i), nil
		}
	}
	return Symbol(0), errors.New("Invalid symbol")
}
