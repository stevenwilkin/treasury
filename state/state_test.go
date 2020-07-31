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

	if s.assets[venue.Nexo][asset.BTC] != 1.2 {
		t.Error("Asset should be set")
	}
}

func TestSetAssetTwiceForVenue(t *testing.T) {
	s := NewState()
	s.SetAsset(venue.Nexo, asset.BTC, 1.2)
	s.SetAsset(venue.Nexo, asset.USDT, 1000)

	if s.assets[venue.Nexo][asset.BTC] != 1.2 {
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

	if s.symbols[symbol.BTCTHB] != 300000 {
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
