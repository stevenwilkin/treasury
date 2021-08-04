package alert

import (
	"fmt"

	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"
)

type direction int

const (
	rising direction = iota
	falling
)

type PriceAlert struct {
	active    bool
	state     *state.State
	symbol    symbol.Symbol
	price     float64
	direction direction
}

func (a *PriceAlert) Description() string {
	return fmt.Sprintf("Price alert at %s %.2f", a.symbol, a.price)
}

func (a *PriceAlert) Message() string {
	return fmt.Sprintf("%s has reached %.2f", a.symbol, a.price)
}

func (a *PriceAlert) Active() bool {
	return a.active
}
func (a *PriceAlert) Deactivate() {
	a.active = false
}

func (a *PriceAlert) Check() bool {
	currentPrice := a.state.Symbol(a.symbol)

	if a.direction == rising {
		return currentPrice >= a.price
	} else {
		return currentPrice <= a.price
	}
}

func NewPriceAlert(s *state.State, sym symbol.Symbol, price float64) *PriceAlert {
	d := rising
	if s.Symbol(sym) > price {
		d = falling
	}

	return &PriceAlert{
		active:    true,
		state:     s,
		symbol:    sym,
		price:     price,
		direction: d}
}

func (a *Alerter) AddPriceAlert(price float64) {
	alert := NewPriceAlert(a.state, symbol.BTCUSDT, price)
	a.AddAlert(alert)
}

var _ Alert = &PriceAlert{}
