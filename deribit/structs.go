package deribit

type authResponse struct {
	Result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	} `json:"result"`
}

type requestMessage struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type portfolioResponse struct {
	Method string `json:"method"`
	Params struct {
		Data struct {
			Equity float64 `json:"equity"`
		} `json:"data"`
	} `json:"params"`
}

type positionsResponse struct {
	Result []struct {
		Size float64 `json:"size"`
	} `json:"result"`
}

type accountSummaryResponse struct {
	Result struct {
		Equity        float64 `json:"equity"`
		InitialMargin float64 `json:"initial_margin"`
	} `json:"result"`
}
