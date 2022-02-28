package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

type State struct {
	mu              sync.Mutex
	Cost            float64
	Assets          map[venue.Venue]map[asset.Asset]float64
	Symbols         map[symbol.Symbol]float64
	FundingRate     [2]float64
	Size            int
	Loan            float64
	FundingAlert    bool
	PriceAlerts     []float64
	LeverageDeribit float64
	LeverageBybit   float64
}

const (
	statePath = "/var/lib/treasuryd/state.json"
)

func NewState() *State {
	return &State{
		Assets:  map[venue.Venue]map[asset.Asset]float64{},
		Symbols: map[symbol.Symbol]float64{}}
}

func (s *State) SetAsset(v venue.Venue, a asset.Asset, q float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Assets[v]; !ok {
		s.Assets[v] = map[asset.Asset]float64{}
	}

	s.Assets[v][a] = q
}

func (s *State) GetAsset(v venue.Venue, a asset.Asset) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Assets[v][a]
}

func (s *State) GetAssets() map[venue.Venue]map[asset.Asset]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := map[venue.Venue]map[asset.Asset]float64{}

	for v, balances := range s.Assets {
		venueAssets := map[asset.Asset]float64{}

		for a, q := range balances {
			if strings.HasPrefix(a.String(), "USD") && q < 0.01 {
				continue
			}

			if a == asset.BTC && q <= 0.00000100 {
				continue
			}

			if q > 0 {
				venueAssets[a] = q
			}
		}

		if len(venueAssets) > 0 {
			results[v] = venueAssets
		}
	}

	return results
}

func (s *State) SetSymbol(sym symbol.Symbol, v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Symbols[sym] = v
}

func (s *State) Symbol(sym symbol.Symbol) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Symbols[sym]
}

func (s *State) GetSymbols() map[symbol.Symbol]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := map[symbol.Symbol]float64{}

	for sym, p := range s.Symbols {
		results[sym] = p
	}

	return results
}

func (s *State) SetCost(c float64) {
	s.Cost = c
}

func (s *State) SetFundingRate(current, predicted float64) {
	s.FundingRate = [2]float64{current, predicted}
}

func (s *State) GetFundingRate() (float64, float64) {
	return s.FundingRate[0], s.FundingRate[1]
}

func (s *State) SetSize(size int) {
	s.Size = size
}

func (s *State) GetLoan() float64 {
	return s.Loan
}

func (s *State) SetLoan(loan float64) {
	s.Loan = loan
}

func (s *State) GetLeverageDeribit() float64 {
	return s.LeverageDeribit
}

func (s *State) SetLeverageDeribit(leverage float64) {
	s.LeverageDeribit = leverage
}
func (s *State) GetLeverageBybit() float64 {
	return s.LeverageBybit
}

func (s *State) SetLeverageBybit(leverage float64) {
	s.LeverageBybit = leverage
}

func (s *State) GetSize() int {
	return s.Size
}

func (s *State) TotalValue() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := 0.0

	for _, balances := range s.Assets {
		for a, quantity := range balances {
			sym, err := symbol.FromString(fmt.Sprintf("%sTHB", a))
			if err == nil {
				total += quantity * s.Symbols[sym]
			}
		}
	}

	if s.Loan > 0 {
		total -= (s.Loan * s.Symbols[symbol.USDTHB])
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

func (s *State) TotalEquity() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := 0.0

	for _, balances := range s.Assets {
		total += balances[asset.BTC]
	}

	return total
}

func (s *State) Exposure() float64 {
	equivalentEquity := float64(s.GetSize()) / s.Symbol(symbol.BTCUSDT)
	return s.TotalEquity() - equivalentEquity
}

func (s *State) THBPremium() float64 {
	btcthb := s.Symbol(symbol.BTCTHB)
	btcusdt := s.Symbol(symbol.BTCUSDT)
	usdtthb := s.Symbol(symbol.USDTTHB)

	if !(btcthb > 0 && btcusdt > 0 && usdtthb > 0) {
		return 0
	}

	equivalent := btcthb / usdtthb
	difference := equivalent - btcusdt
	percentage := difference / btcusdt

	return percentage
}

func (s *State) USDTPremium() float64 {
	usdthb := s.Symbol(symbol.USDTHB)
	usdtthb := s.Symbol(symbol.USDTTHB)

	if !(usdthb > 0 && usdtthb > 0) {
		return 0
	}

	return (usdtthb - usdthb) / usdthb
}

func (s *State) GetPriceAlerts() []float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.PriceAlerts
}

func (s *State) SetPriceAlerts(alerts []float64) {
	s.PriceAlerts = alerts
}

func (s *State) GetFundingAlert() bool {
	return s.FundingAlert
}

func (s *State) SetFundingAlert(alert bool) {
	s.FundingAlert = alert
}

func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
