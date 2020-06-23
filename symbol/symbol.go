package symbol

type Symbol int

const (
	BTCTHB Symbol = iota
	USDTTHB
	USDTHB
)

type Prices map[Symbol]float64
