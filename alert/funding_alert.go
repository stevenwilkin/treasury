package alert

import (
	"fmt"

	"github.com/stevenwilkin/treasury/state"
)

type FundingAlert struct {
	active bool
	state  *state.State
}

func (a *FundingAlert) Description() string {
	return "Negative funding alert"
}

func (a *FundingAlert) Message() string {
	current, expected := a.state.Funding()

	return fmt.Sprintf(
		"Funding: %f%%, Predicted: %f%%", current*100, expected*100)
}

func (a *FundingAlert) Active() bool {
	return a.active
}

func (a *FundingAlert) Deactivate() {
	a.active = false
}

func (a *FundingAlert) Check() bool {
	current, expected := a.state.Funding()

	return current < 0 || expected < 0
}

func NewFundingAlert(s *state.State) *FundingAlert {
	return &FundingAlert{
		active: true,
		state:  s}
}

var _ Alert = &FundingAlert{}
