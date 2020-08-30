package ftx

type walletResponse struct {
	Result map[string][]struct {
		Coin  string
		Total float64
	}
}

type orderRequest struct {
	Market   string  `json:"market"`
	Side     string  `json:"side"`
	Size     float64 `json:"size"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
	PostOnly bool    `json:"postOnly"`
}

type editOrderRequest struct {
	Size  float64 `json:"size"`
	Price float64 `json:"price"`
}

type orderResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Id int64 `json:id`
	} `json:"result"`
}

type opMessage struct {
	Op   string                 `json:"op"`
	Args map[string]interface{} `json:"args"`
}

type tradeMessage struct {
	Channel string `json:"channel"`
	Market  string `json:"market"`
	Type    string `json:"type"`
	Data    struct {
		Bid        float64 `json:"bid"`
		Ask        float64 `json:"ask"`
		Id         int64   `json:id`
		Status     string  `json:"status"`
		Size       float64 `json:size`
		FilledSize float64 `json:filledSize`
	} `json:"data"`
}
