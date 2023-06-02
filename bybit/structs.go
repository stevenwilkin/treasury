package bybit

type fundingResponse struct {
	Result struct {
		List []struct {
			FundingRate string `json:"fundingRate"`
		} `json:"list"`
	} `json:"result"`
}

type positionResponse struct {
	Result struct {
		List []struct {
			Size          string `json:"size"`
			PositionValue string `json:"positionValue"`
		} `json:"list"`
	} `json:"result"`
}

type walletResponse struct {
	Result struct {
		List []struct {
			Coin []struct {
				Equity string `json:"equity"`
			} `json:"coin"`
		} `json:"list"`
	} `json:"result"`
}
