package state

import (
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

type State struct {
	assets  map[venue.Venue]map[asset.Asset]float64
	symbols map[symbol.Symbol]float64
}

func NewState() *State {
	return &State{
		assets:  map[venue.Venue]map[asset.Asset]float64{},
		symbols: map[symbol.Symbol]float64{}}
}

func (s *State) SetAsset(v venue.Venue, a asset.Asset, q float64) {
	if _, ok := s.assets[v]; !ok {
		s.assets[v] = map[asset.Asset]float64{}
	}

	s.assets[v][a] = q
}

func (s *State) Asset(v venue.Venue, a asset.Asset) float64 {
	return s.assets[v][a]
}

func (s *State) SetSymbol(sym symbol.Symbol, v float64) {
	s.symbols[sym] = v
}

func (s *State) Symbol(sym symbol.Symbol) float64 {
	return s.symbols[sym]
}
