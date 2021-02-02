package feed

import (
	"errors"
	"strings"
)

type Feed int

const (
	BTCUSDT Feed = iota
	BTCTHB
	USDTTHB
	USDTHB
	Binance
	Deribit
	Bybit
	FTX
	Funding
)

func feeds() []string {
	return []string{
		"BTCUSDT",
		"BTCTHB",
		"USDTTHB",
		"USDTHB",
		"Binance",
		"Deribit",
		"Bybit",
		"FTX",
		"Funding"}
}

func (s Feed) String() string {
	return feeds()[s]
}

func FromString(s string) (Feed, error) {
	for i, feed := range feeds() {
		if strings.ToLower(s) == strings.ToLower(feed) {
			return Feed(i), nil
		}
	}
	return Feed(0), errors.New("Invalid feed")
}
