package venue

import (
	"errors"
	"strings"
)

type Venue int

const (
	FTX Venue = iota
	Nexo
	Blockfi
	Hodlnaut
	Deribit
	Bybit
	Binance
)

func venues() []string {
	return []string{"FTX", "Nexo", "Blockfi", "Hodlnaut", "Deribit", "Bybit", "Binance"}
}

func (v Venue) String() string {
	return venues()[v]
}

func Exists(name string) bool {
	for _, venue := range venues() {
		if strings.ToLower(name) == strings.ToLower(venue) {
			return true
		}
	}
	return false
}

func FromString(s string) (Venue, error) {
	for i, venue := range venues() {
		if strings.ToLower(s) == strings.ToLower(venue) {
			return Venue(i), nil
		}
	}
	return Venue(0), errors.New("Invalid venue")
}
