package bybit

import (
	"encoding/json"
)

type equityResponse struct {
	Result struct {
		BTC struct {
			Equity float64
		}
	}
}

type fundingResponse struct {
	Result []struct {
		FundingRate          string `json:"funding_rate"`
		PredictedFundingRate string `json:"predicted_funding_rate"`
	} `json:"result"`
}

type positionResponse struct {
	Result struct {
		Size int `json:"size"`
	} `json:"result"`
}

type wsCommand struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

type wsResponse struct {
	Topic string          `json:"topic"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}

type order struct {
	Id    int64  `json:"id"`
	Price string `json:"price"`
	Side  string `json:"side"`
}
type snapshotData []order

type updateData struct {
	Delete []order `json:"delete"`
	Insert []order `json:"insert"`
}

type orderResponse struct {
	Result struct {
		OrderId string `json:"order_id"`
	} `json:"result"`
}

type orderTopicData struct {
	Topic string `json:"topic"`
	Data  []struct {
		OrderId     string `json:"order_id"`
		OrderStatus string `json:"order_status"`
		Price       string `json:"price"`
		Qty         int    `json:"qty"`
		CumExecQty  int    `json:"cum_exec_qty"`
	} `json:"data"`
}
