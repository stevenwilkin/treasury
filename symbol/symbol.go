package symbol

type Symbol int

const (
	BTCTHB Symbol = iota
	USDTTHB
)

type Prices map[Symbol]float64
