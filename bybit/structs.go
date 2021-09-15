package bybit

type fundingResponse struct {
	Result []struct {
		FundingRate          string `json:"funding_rate"`
		PredictedFundingRate string `json:"predicted_funding_rate"`
	} `json:"result"`
}

type positionResponse struct {
	Result struct {
		Size          int     `json:"size"`
		PositionValue string  `json:"position_value"`
		WalletBalance string  `json:"wallet_balance"`
		UnrealisedPnl float64 `json:"unrealised_pnl"`
	} `json:"result"`
}
