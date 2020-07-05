package symbol

type Symbol int

const (
	BTCTHB Symbol = iota
	USDTTHB
	USDTHB
)

type Prices map[Symbol]float64

func (s Symbol) String() string {
	return []string{"BTCTHB", "USDTTHB", "USDTHB"}[s]
}
