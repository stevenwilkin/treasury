package bybit

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
