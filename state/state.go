package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

type State struct {
	symbolSubscribers map[chan SymbolNotification]bool
	Cost              float64
	Assets            map[venue.Venue]map[asset.Asset]float64
	Symbols           map[symbol.Symbol]float64
}

const (
	statePath = "/tmp/treasuryd.json"
)

func NewState() *State {
	return &State{
		symbolSubscribers: map[chan SymbolNotification]bool{},
		Assets:            map[venue.Venue]map[asset.Asset]float64{},
		Symbols:           map[symbol.Symbol]float64{}}
}

func (s *State) SetAsset(v venue.Venue, a asset.Asset, q float64) {
	if _, ok := s.Assets[v]; !ok {
		s.Assets[v] = map[asset.Asset]float64{}
	}

	s.Assets[v][a] = q
}

func (s *State) Asset(v venue.Venue, a asset.Asset) float64 {
	return s.Assets[v][a]
}

func (s *State) SetSymbol(sym symbol.Symbol, v float64) {
	old := s.Symbols[sym]

	s.Symbols[sym] = v

	if old != v {
		s.NotifySymbolSubscribers(sym, v)
	}
}

func (s *State) Symbol(sym symbol.Symbol) float64 {
	return s.Symbols[sym]
}

func (s *State) SetCost(c float64) {
	s.Cost = c
}

func (s *State) TotalValue() float64 {
	total := 0.0

	for _, balances := range s.Assets {
		for a, quantity := range balances {
			sym, err := symbol.FromString(fmt.Sprintf("%sTHB", a))
			if err == nil {
				total += quantity * s.Symbols[sym]
			}
		}
	}

	return total
}

func (s *State) Pnl() float64 {
	return s.TotalValue() - s.Cost
}

func (s *State) PnlPercentage() float64 {
	if s.Cost == 0 {
		return 0
	}

	return (s.Pnl() / s.Cost) * 100
}

func (s *State) Save() error {
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile("/tmp/", "state")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(tmpFile.Name(), b, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), statePath)
	if err != nil {
		os.RemoveAll(tmpFile.Name())
		return err
	}

	return nil
}

func (s *State) Load() error {
	stateJSON, err := ioutil.ReadFile(statePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(stateJSON), s)
	if err != nil {
		return err
	}

	return nil
}
