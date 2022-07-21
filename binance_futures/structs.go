package binance_futures

type errorResponse struct {
	Msg string `json:"msg"`
}

type balanceResponse []struct {
	Asset   string `json:"asset"`
	Balance string `json:"balance"`
}

type accountResponse struct {
	Positions []struct {
		PositionAmt string `json:"positionAmt"`
	} `json:"positions"`
}
