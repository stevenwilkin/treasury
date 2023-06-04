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
	funding := a.state.GetFundingRate()

	return fmt.Sprintf("Funding: %f%%", funding*100)
}

func (a *FundingAlert) Active() bool {
	return a.active
}

func (a *FundingAlert) Deactivate() {
	a.active = false
}

func (a *FundingAlert) Check() bool {
	funding := a.state.GetFundingRate()

	return funding < 0
}

func NewFundingAlert(s *state.State) *FundingAlert {
	return &FundingAlert{
		active: true,
		state:  s}
}

func (a *Alerter) AddFundingAlert() {
	alert := NewFundingAlert(a.state)
	a.AddAlert(alert)
}

var _ Alert = &FundingAlert{}
