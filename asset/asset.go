package asset

import (
	"errors"
	"strings"
)

type Asset int

const (
	BTC Asset = iota
	USDT
	USDC
	USD
)

type Balances map[Asset]float64

func assets() []string {
	return []string{"BTC", "USDT", "USDC", "USD"}
}

func (a Asset) String() string {
	return assets()[a]
}

func FromString(s string) (Asset, error) {
	for i, asset := range assets() {
		if strings.ToLower(s) == strings.ToLower(asset) {
			return Asset(i), nil
		}
	}
	return Asset(0), errors.New("Invalid asset")
}
