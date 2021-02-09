package bitkub

type tickerMessage struct {
	Last float64 `json:"last"`
}

type tickerResponse map[string]struct {
	Last float64 `json:"last"`
}
