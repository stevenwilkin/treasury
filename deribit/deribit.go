package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type Deribit struct {
	ApiId        string
	ApiSecret    string
	Test         bool
	_accessToken string
}

func (d *Deribit) accessToken() string {
	if d._accessToken != "" {
		return d._accessToken
	}

	log.WithField("venue", "deribit").Debug("Fetching access token")

	v := url.Values{}
	v.Set("client_id", d.ApiId)
	v.Set("client_secret", d.ApiSecret)
	v.Set("grant_type", "client_credentials")

	u := url.URL{
		Scheme:   "https",
		Host:     d.hostname(),
		Path:     "/api/v2/public/auth",
		RawQuery: v.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response authResponse
	json.Unmarshal(body, &response)
	d._accessToken = response.Result.AccessToken

	return d._accessToken
}

func (d *Deribit) hostname() string {
	if d.Test {
		return "test.deribit.com"
	} else {
		return "www.deribit.com"
	}
}

func (d *Deribit) subscribe(channels []string) *websocket.Conn {
	socketUrl := url.URL{Scheme: "wss", Host: d.hostname(), Path: "/ws/api/v2"}

	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	request := requestMessage{
		Method: "/private/subscribe",
		Params: map[string]interface{}{
			"channels":     channels,
			"access_token": d.accessToken()}}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		log.Panic(err.Error())
	}

	if err = c.WriteMessage(websocket.TextMessage, jsonRequest); err != nil {
		log.Panic(err.Error())
	}

	return c
}

func (d *Deribit) Equity() chan float64 {
	log.WithField("venue", "deribit").Info("Subscribing to equity")

	c := d.subscribe([]string{"user.portfolio.BTC"})
	ch := make(chan float64)

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "deribit").Info("Reconnecting to equity subscription")
				c.Close()
				d._accessToken = "" // force fresh access token
				c = d.subscribe([]string{"user.portfolio.BTC"})
				continue
			}

			var response portfolioResponse
			json.Unmarshal(message, &response)

			if response.Method == "" {
				continue
			}

			log.WithFields(log.Fields{
				"venue": "deribit",
				"asset": "BTC",
				"value": response.Params.Data.Equity,
			}).Debug("Received equity")
			ch <- response.Params.Data.Equity
		}
	}()

	return ch
}

func (d *Deribit) GetSize() int {
	u := "https://www.deribit.com/api/v2/private/get_positions?currency=BTC&kind=future"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response positionsResponse
	json.Unmarshal(body, &response)

	size := 0.0
	for _, position := range response.Result {
		size += position.Size
	}

	return int(math.Abs(size))
}

func (d *Deribit) PlaceOrder(instrument string, amount int, price float64, buy, reduce bool) string {
	log.WithFields(log.Fields{
		"venue":      "deribit",
		"instrument": instrument,
		"amount":     amount,
		"price":      price,
	}).Info("Placing order")

	v := url.Values{}
	v.Set("instrument_name", instrument)
	v.Set("amount", strconv.Itoa(amount))
	v.Set("price", strconv.FormatFloat(price, 'f', 2, 64))
	v.Set("post_only", "true")
	v.Set("reject_post_only", "true")

	path := "/api/v2/private/buy"
	if !buy {
		path = "/api/v2/private/sell"
	}

	if reduce {
		v.Set("reduce_only", "true")
	}

	u := url.URL{
		Scheme:   "https",
		Host:     d.hostname(),
		Path:     path,
		RawQuery: v.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}

	var response orderResponse
	json.Unmarshal(body, &response)

	return response.Result.Order.OrderId
}

func (d *Deribit) EditOrder(orderId string, amount int, price float64, reduce bool) {
	log.WithFields(log.Fields{
		"venue":  "deribit",
		"order":  orderId,
		"amount": amount,
		"price":  price,
	}).Info("Updating order")

	v := url.Values{}
	v.Set("order_id", orderId)
	v.Set("amount", strconv.Itoa(amount))
	v.Set("price", strconv.FormatFloat(price, 'f', 2, 64))
	v.Set("post_only", "true")
	v.Set("reject_post_only", "true")

	if reduce {
		v.Set("reduce_only", "true")
	}

	u := url.URL{
		Scheme:   "https",
		Host:     d.hostname(),
		Path:     "/api/v2/private/edit",
		RawQuery: v.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessToken()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		log.Panic(err.Error())
	}
}

func (d *Deribit) Trade(instrument string, contracts int, buy, reduce bool) {
	var bestPrice, price float64
	var orderId string

	ordersChannel := fmt.Sprintf("user.orders.%s.raw", instrument)
	quoteChannel := fmt.Sprintf("quote.%s", instrument)

	c := d.subscribe([]string{ordersChannel, quoteChannel})
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}

		var response tradeResponse
		json.Unmarshal(message, &response)

		if response.Method != "subscription" {
			continue
		}

		switch response.Params.Channel {
		case quoteChannel:
			if buy {
				bestPrice = response.Params.Data.BestBidPrice
			} else {
				bestPrice = response.Params.Data.BestAskPrice
			}

			if orderId == "" {
				price = bestPrice
				orderId = d.PlaceOrder(instrument, contracts, price, buy, reduce)
			} else if price != bestPrice {
				price = bestPrice
				d.EditOrder(orderId, contracts, price, reduce)
			}
		case ordersChannel:
			switch response.Params.Data.OrderState {
			case "open":
				log.WithFields(log.Fields{
					"venue":    "deribit",
					"order":    orderId,
					"quantity": response.Params.Data.FilledAmount,
				}).Debug("Fill")
			case "cancelled":
				log.WithFields(log.Fields{
					"venue": "deribit",
					"order": orderId,
				}).Debug("Order cancelled")
				return
			case "filled":
				log.WithFields(log.Fields{
					"venue": "deribit",
					"order": orderId,
				}).Info("Order filled")
				return
			}
		}
	}
}
