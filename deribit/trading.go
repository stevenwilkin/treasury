package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

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
