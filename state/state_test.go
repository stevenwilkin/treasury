package state

import (
	"testing"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

func TestSetAsset(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1.2)

	if s.Assets[venue.Nexo][asset.BTC] != 1.2 {
		t.Error("Asset should be set")
	}
}

func TestSetAssetTwiceForVenue(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1.2)
	s.SetAsset(venue.Nexo, asset.USDT, 1000)

	if s.Assets[venue.Nexo][asset.BTC] != 1.2 {
		t.Error("Asset quantity should not be overwritten")
	}
}

func TestAsset(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1.2)

	if s.Asset(venue.Nexo, asset.BTC) != 1.2 {
		t.Error("Asset quantity should be returned")
	}
}

func TestSetSymbol(t *testing.T) {
	s := NewState()
	s.SetSymbol(symbol.BTCTHB, 300000)

	if s.Symbols[symbol.BTCTHB] != 300000 {
		t.Error("Symbol should be set")
	}
}

func TestSymbol(t *testing.T) {
	s := NewState()
	s.SetSymbol(symbol.BTCTHB, 300000)

	if s.Symbol(symbol.BTCTHB) != 300000 {
		t.Error("Symbol value should be returned")
	}
}

func TestFunding(t *testing.T) {
	s := NewState()
	s.SetFunding(1.1, 2.2)

	current, predicted := s.Funding()

	if current != 1.1 || predicted != 2.2 {
		t.Errorf("Expected: 1.1, 2.2 - Got: %f, %f", current, predicted)
	}
}

func TestSize(t *testing.T) {
	s := NewState()
	s.SetSize(10000)

	if s.Size() != 10000 {
		t.Errorf("Expected: 10000 - Got: %d", s.Size())
	}
}

func TestTotalValue(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1.1)
	s.SetAsset(venue.Nexo, asset.USDT, 1000)
	s.SetAsset(venue.FTX, asset.BTC, 0.1)
	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetSymbol(symbol.USDTTHB, 31)

	if s.TotalValue() != 391000 {
		t.Errorf("Expected total value to be %d, got %f", 391000, s.TotalValue())
	}
}

func TestTotalEquity(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1)
	s.SetAsset(venue.Nexo, asset.USDT, 1000)
	s.SetAsset(venue.FTX, asset.BTC, 2)

	if s.TotalEquity() != 3 {
		t.Errorf("Expected total equity to be %f, got %f", 3.0, s.TotalEquity())
	}
}

func TestPnl(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1)
	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetCost(200000)

	if s.Pnl() != 100000 {
		t.Errorf("Expected PnL to be %d, got %f", 100000, s.Pnl())
	}
}

func TestPnlPercentageWhenNoCost(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1)
	s.SetSymbol(symbol.BTCTHB, 300000)
	s.SetCost(0)

	if s.PnlPercentage() != 0 {
		t.Errorf("Expected PnL %% to be %d, got %f", 0, s.PnlPercentage())
	}
}

func TestPnlPercentage(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1)
	s.SetSymbol(symbol.BTCTHB, 220000)
	s.SetCost(200000)

	if s.PnlPercentage() != 10 {
		t.Errorf("Expected PnL %% to be %d, got %f", 10, s.PnlPercentage())
	}
}

func TestExposure(t *testing.T) {
	s := NewState()
	s.SetSize(90000)
	s.SetAsset(venue.Nexo, asset.BTC, 10)
	s.SetAsset(venue.Nexo, asset.USDT, 1000)
	s.SetSymbol(symbol.BTCUSDT, 10000)

	if s.Exposure() != 1 {
		t.Errorf("Expected exposure to be %f, got %f", 1.0, s.Exposure())
	}
}

func TestTHBPremium(t *testing.T) {
	s := NewState()

	if s.THBPremium() != 0 {
		t.Error("Expected THB premium to be 0")
	}

	s.SetSymbol(symbol.BTCTHB, 330000)

	if s.THBPremium() != 0 {
		t.Error("Expected THB premium to be 0")
	}

	s.SetSymbol(symbol.BTCUSDT, 10000)

	if s.THBPremium() != 0 {
		t.Error("Expected THB premium to be 0")
	}

	s.SetSymbol(symbol.USDTTHB, 30)

	if s.THBPremium() != 0.1 {
		t.Errorf("Expected THB premium to be %f, got %f", 0.1, s.THBPremium())
	}
}

func TestUSDTPremium(t *testing.T) {
	s := NewState()

	if s.USDTPremium() != 0 {
		t.Error("Expected USDT premium to be 0")
	}

	s.SetSymbol(symbol.USDTHB, 30)

	if s.USDTPremium() != 0 {
		t.Error("Expected USDT premium to be 0")
	}

	s.SetSymbol(symbol.USDTTHB, 33)

	if s.USDTPremium() != 0.1 {
		t.Errorf("Expected USDT premium to be %f, got %f", 0.1, s.USDTPremium())
	}
}
