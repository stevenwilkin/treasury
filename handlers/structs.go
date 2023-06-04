package handlers

import "time"

type pricesMessage struct {
	Prices map[string]float64 `json:"prices"`
}

type assetsMessage struct {
	Assets map[string]map[string]float64 `json:"assets"`
}

type pnlMessage struct {
	Cost          float64 `json:"cost"`
	Value         float64 `json:"value"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
}

type alertMessage struct {
	Active      bool   `json:"active"`
	Description string `json:"description"`
}

type fundingMessage struct {
	Value float64 `float64:"value"`
}

type feedsResponseItem struct {
	Active     bool
	LastUpdate time.Time
}
type feedsResponse struct {
	Feeds map[string]feedsResponseItem
}
