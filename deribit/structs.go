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

type quoteMessage struct {
	Method string `json:"method"`
	Params struct {
		Data struct {
			BestBidPrice float64 `json:"best_bid_price"`
			BestAskPrice float64 `json:"best_ask_price"`
		} `json:"data"`
	} `json:"params"`
}

type orderMessage struct {
	Method string `json:"method"`
	Params struct {
		Data struct {
			OrderId      string  `json:"order_id"`
			OrderState   string  `json:"order_state"`
			FilledAmount float64 `json:"filled_amount"`
		} `json:"data"`
	} `json:"params"`
}

type orderResponse struct {
	Result struct {
		Order struct {
			OrderId string `json:"order_id"`
		} `json:"order"`
	} `json:"result"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
