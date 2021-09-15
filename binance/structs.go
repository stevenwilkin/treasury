package binance

import "strconv"

type tickerMessage struct {
	P string
}

type assetBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

func (ab *assetBalance) Total() float64 {
	free, _ := strconv.ParseFloat(ab.Free, 64)
	locked, _ := strconv.ParseFloat(ab.Locked, 64)
	return free + locked
}

type accountResponse struct {
	Balances []assetBalance `json:"balances"`
}
