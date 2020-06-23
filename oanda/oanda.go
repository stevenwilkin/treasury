package oanda

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/stevenwilkin/treasury/symbol"
)

type Oanda struct {
	ApiKey    string
	AccountId string
}

type priceResponse struct {
	Prices []struct {
		Bids []struct {
			Price string
		}
		Asks []struct {
			Price string
		}
	}
}

func (o *Oanda) Price(s symbol.Symbol) float64 {
	var ticker string

	switch s {
	case symbol.USDTHB:
		ticker = "USD_THB"
	}

	url := fmt.Sprintf(
		"https://api-fxtrade.oanda.com/v3/accounts/%s/pricing?instruments=%s",
		o.AccountId,
		ticker)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	var response priceResponse
	json.Unmarshal(body, &response)

	bidString := response.Prices[0].Bids[0].Price
	askString := response.Prices[0].Asks[0].Price

	bid, _ := strconv.ParseFloat(bidString, 64)
	ask, _ := strconv.ParseFloat(askString, 64)
	price := (bid + ask) / 2

	return price
}
