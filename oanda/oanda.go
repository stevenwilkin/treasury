package oanda

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/stevenwilkin/treasury/symbol"

	log "github.com/sirupsen/logrus"
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

func (o *Oanda) GetPrice(s symbol.Symbol) (float64, error) {
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
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response priceResponse
	json.Unmarshal(body, &response)

	if len(response.Prices) != 1 ||
		len(response.Prices[0].Bids) != 1 ||
		len(response.Prices[0].Asks) != 1 {
		return 0, errors.New("Invalid price response")
	}

	bidString := response.Prices[0].Bids[0].Price
	askString := response.Prices[0].Asks[0].Price

	bid, _ := strconv.ParseFloat(bidString, 64)
	ask, _ := strconv.ParseFloat(askString, 64)
	price := (bid + ask) / 2

	return price, nil
}

func (o *Oanda) Price(s symbol.Symbol) chan float64 {
	log.WithFields(log.Fields{
		"venue":  "oanda",
		"symbol": s,
	}).Info("Polling price")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			price, err := o.GetPrice(s)
			if err != nil {
				log.WithField("venue", "oanda").Error(err.Error())
				<-ticker.C
				continue
			}

			log.WithFields(log.Fields{
				"venue":  "oanda",
				"symbol": s,
				"value":  price,
			}).Debug("Received price")

			ch <- price
			<-ticker.C
		}
	}()

	return ch
}
