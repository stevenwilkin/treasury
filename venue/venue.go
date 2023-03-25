package venue

import (
	"errors"
	"strings"
)

type Venue int

const (
	Nexo Venue = iota
	Blockfi
	Hodlnaut
	Deribit
	Bybit
	Binance
	Ledn
	Wasabi
	Celsius
	Ledger
)

func venues() []string {
	return []string{"Nexo", "Blockfi", "Hodlnaut", "Deribit", "Bybit", "Binance", "Ledn", "Wasabi", "Celsius", "Ledger"}
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
