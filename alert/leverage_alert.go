package alert

import (
	"fmt"

	"github.com/stevenwilkin/treasury/state"
)

type LeverageAlert struct {
	active    bool
	state     *state.State
	threshold float64
}

func (a *LeverageAlert) Description() string {
	return fmt.Sprintf("Leverage alert at %.2f", a.threshold)
}

func (a *LeverageAlert) Message() string {
	return fmt.Sprintf("Leverage: %.2f %.2f",
		a.state.GetLeverageDeribit(), a.state.GetLeverageBybit())
}

func (a *LeverageAlert) Active() bool {
	return a.active
}

func (a *LeverageAlert) Deactivate() {
	a.active = false
}

func (a *LeverageAlert) Check() bool {
	return a.state.GetLeverageDeribit() >= a.threshold ||
		a.state.GetLeverageBybit() >= a.threshold
}

func NewLeverageAlert(s *state.State, threshold float64) *LeverageAlert {
	return &LeverageAlert{
		active:    true,
		state:     s,
		threshold: threshold}
}

var _ Alert = &LeverageAlert{}
