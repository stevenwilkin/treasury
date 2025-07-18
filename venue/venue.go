package venue

import (
	"errors"
	"strings"
)

type Venue int

const (
	Nexo Venue = iota
	Deribit
	Bybit
	Binance
	Ledn
	Loan
	Ledger
)

func venues() []string {
	return []string{"Nexo", "Deribit", "Bybit", "Binance", "Ledn", "Loan", "Ledger"}
}

func (v Venue) String() string {
	return venues()[v]
}

func FromString(s string) (Venue, error) {
	for i, venue := range venues() {
		if strings.ToLower(s) == strings.ToLower(venue) {
			return Venue(i), nil
		}
	}
	return Venue(0), errors.New("Invalid venue")
}
