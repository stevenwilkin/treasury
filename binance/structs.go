package binance

type tickerMessage struct {
	P string
}

type accountResponse struct {
	Balances []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
}

type listenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

type bookTickerMessage struct {
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

type createOrderResponse struct {
	OrderId int64 `json:"orderId"`
}

type userDataMessage struct {
	EventType   string `json:"e"`
	EventTime   int64  `json:"E"`
	OrderStatus string `json:"X"`
	OrderId     int64  `json:"i"`
	Ignore      int64  `json:"I"`
	FillQty     string `json:"l"`
	FillPrice   string `json:"L"`
	CumFillQty  string `json:"z"`
	CumQuoteQty string `json:"Z"`
}
