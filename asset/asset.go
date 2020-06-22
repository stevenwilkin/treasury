package asset

type Asset int

const (
	BTC Asset = iota
	USDT
	USD
)

type Balances map[Asset]float64
