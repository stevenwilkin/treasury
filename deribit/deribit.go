package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (d *Deribit) accessToken() (string, error) {
	if d._accessToken != "" {
		return d._accessToken, nil
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
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response authResponse
	json.Unmarshal(body, &response)
	d._accessToken = response.Result.AccessToken

	return d._accessToken, nil
}

func (d *Deribit) hostname() string {
	if d.Test {
		return "test.deribit.com"
	} else {
		return "www.deribit.com"
	}
}

func (d *Deribit) subscribe(channels []string) (*websocket.Conn, error) {
	socketUrl := url.URL{Scheme: "wss", Host: d.hostname(), Path: "/ws/api/v2"}

	c, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	authRequest := requestMessage{
		Method: "/public/auth",
		Params: map[string]interface{}{
			"client_id":     d.ApiId,
			"client_secret": d.ApiSecret,
			"grant_type":    "client_credentials"}}

	if err = c.WriteJSON(authRequest); err != nil {
		return nil, err
	}

	request := requestMessage{
		Method: "/private/subscribe",
		Params: map[string]interface{}{
			"channels": channels}}

	if err = c.WriteJSON(request); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(10 * time.Second)
	testMessage := requestMessage{Method: "/public/test"}

	go func() {
		for {
			if err = c.WriteJSON(testMessage); err != nil {
				log.Error(err.Error())
			}
			<-ticker.C
		}
	}()

	return c, nil
}

func (d *Deribit) Equity() chan float64 {
	log.WithField("venue", "deribit").Info("Subscribing to equity")

	ch := make(chan float64)
	c, err := d.subscribe([]string{"user.portfolio.BTC"})
	if err != nil {
		log.Error(err.Error())
		return ch
	}

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "deribit").Info("Reconnecting to equity subscription")
				c.Close()
				c, _ = d.subscribe([]string{"user.portfolio.BTC"})
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
		log.Error(err.Error())
		return 0
	}

	accessToken, err := d.accessToken()
	if err != nil {
		log.Error(err.Error())
		return 0
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return 0
	}

	var response positionsResponse
	json.Unmarshal(body, &response)

	size := 0.0
	for _, position := range response.Result {
		size += position.Size
	}

	return int(math.Abs(size))
}

func (d *Deribit) PlaceOrder(instrument string, amount int, price float64, buy, reduce bool) (string, error) {
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
		return "", err
	}

	accessToken, err := d.accessToken()
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response orderResponse
	json.Unmarshal(body, &response)

	return response.Result.Order.OrderId, nil
}

func (d *Deribit) EditOrder(orderId string, amount int, price float64, reduce bool) error {
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
		return err
	}

	accessToken, err := d.accessToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	return err
}

func (d *Deribit) Trade(instrument string, contracts int, buy, reduce bool) {
	var bestPrice, price float64
	var orderId string

	ordersChannel := fmt.Sprintf("user.orders.%s.raw", instrument)
	quoteChannel := fmt.Sprintf("quote.%s", instrument)

	c, err := d.subscribe([]string{ordersChannel, quoteChannel})
	if err != nil {
		log.Error(err.Error())
		return
	}
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
				orderId, err = d.PlaceOrder(instrument, contracts, price, buy, reduce)
				if err != nil {
					return
				}
			} else if price != bestPrice {
				price = bestPrice
				err = d.EditOrder(orderId, contracts, price, reduce)
				if err != nil {
					return
				}
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
