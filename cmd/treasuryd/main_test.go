package main

import (
	"testing"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"
)

func TestTotalValue(t *testing.T) {
	assets = map[venue.Venue]map[asset.Asset]float64{
		venue.Nexo: asset.Balances{
			asset.BTC:  1.1,
			asset.USDT: 1000,
		},
		venue.FTX: asset.Balances{
			asset.BTC: 0.1,
		},
	}

	prices = symbol.Prices{
		symbol.BTCTHB:  300000,
		symbol.USDTTHB: 31,
	}

	if totalValue() != 391000 {
		t.Errorf("Expected total value to be %d, got %f", 391000, totalValue())
	}
}

func TestPnl(t *testing.T) {
	cost = 200000

	assets = map[venue.Venue]map[asset.Asset]float64{
		venue.Nexo: asset.Balances{
			asset.BTC: 1,
		},
	}

	prices = symbol.Prices{
		symbol.BTCTHB: 300000,
	}

	if pnl() != 100000 {
		t.Errorf("Expected PnL to be %d, got %f", 100000, pnl())
	}
}

func TestPnlPercentageWhenNoCost(t *testing.T) {
	cost = 0

	assets = map[venue.Venue]map[asset.Asset]float64{
		venue.Nexo: asset.Balances{
			asset.BTC: 1,
		},
	}

	prices = symbol.Prices{
		symbol.BTCTHB: 300000,
	}

	if pnlPercentage() != 0 {
		t.Errorf("Expected PnL %% to be %d, got %f", 0, pnlPercentage())
	}
}

func TestPnlPercentage(t *testing.T) {
	cost = 200000

	assets = map[venue.Venue]map[asset.Asset]float64{
		venue.Nexo: asset.Balances{
			asset.BTC: 1,
		},
	}

	prices = symbol.Prices{
		symbol.BTCTHB: 220000,
	}

	if pnlPercentage() != 10 {
		t.Errorf("Expected PnL %% to be %d, got %f", 10, pnlPercentage())
	}
}
