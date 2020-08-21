package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	log "github.com/sirupsen/logrus"
)

type State struct {
	symbolSubscribers map[chan SymbolNotification]bool
	Cost              float64
	Assets            map[venue.Venue]map[asset.Asset]float64
	Symbols           map[symbol.Symbol]float64
	FundingRate       [2]float64
	TotalSize         int
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

	if s.Assets[v][a] == q {
		return
	}

	log.WithFields(log.Fields{
		"venue":    strings.ToLower(v.String()),
		"asset":    a,
		"quantity": q,
	}).Debug("Updating state")

	s.Assets[v][a] = q
}

func (s *State) Asset(v venue.Venue, a asset.Asset) float64 {
	return s.Assets[v][a]
}

func (s *State) SetSymbol(sym symbol.Symbol, v float64) {
	if s.Symbols[sym] == v {
		return
	}

	log.WithFields(log.Fields{
		"symbol": sym,
		"value":  v,
	}).Debug("Updating state")
	s.Symbols[sym] = v
	s.NotifySymbolSubscribers(sym, v)
}

func (s *State) Symbol(sym symbol.Symbol) float64 {
	return s.Symbols[sym]
}

func (s *State) SetCost(c float64) {
	s.Cost = c
}

func (s *State) SetFunding(current, predicted float64) {
	s.FundingRate = [2]float64{current, predicted}
}

func (s *State) Funding() (float64, float64) {
	return s.FundingRate[0], s.FundingRate[1]
}

func (s *State) SetSize(size int) {
	s.TotalSize = size
}

func (s *State) Size() int {
	return s.TotalSize
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

func (s *State) Exposure() float64 {
	usdt := 0.0

	for _, balances := range s.Assets {
		for a, quantity := range balances {
			if a == asset.USDT {
				usdt += quantity
			}
		}
	}

	dollarExposure := float64(s.Size()) + usdt
	totalValueInDollars := s.TotalValue() / s.Symbol(symbol.USDTHB)
	difference := totalValueInDollars - dollarExposure

	return difference / s.Symbol(symbol.BTCUSDT)
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
