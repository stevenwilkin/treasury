package state

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

type State struct {
	Assets  map[venue.Venue]map[asset.Asset]float64
	Symbols map[symbol.Symbol]float64
}

const (
	statePath = "/tmp/treasuryd.json"
)

func NewState() *State {
	return &State{
		Assets:  map[venue.Venue]map[asset.Asset]float64{},
		Symbols: map[symbol.Symbol]float64{}}
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
	s.Symbols[sym] = v
}

func (s *State) Symbol(sym symbol.Symbol) float64 {
	return s.Symbols[sym]
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
