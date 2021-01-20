package deribit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func (d *Deribit) PlaceOrder(instrument string, amount int, price float64, buy, reduce bool) (string, error) {
	log.WithFields(log.Fields{
		"venue":      "deribit",
		"instrument": instrument,
		"amount":     amount,
		"price":      price,
	}).Debug("Placing order")

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

func (d *Deribit) orderStatus(instrument string) chan bool {
	done := make(chan bool, 1)
	ordersChannel := fmt.Sprintf("user.orders.%s.raw", instrument)

	c, err := d.subscribe([]string{ordersChannel})
	if err != nil {
		log.Error(err.Error())
		return done
	}

	go func() {
		var om orderMessage
		defer c.Close()

		for {
			if err = c.ReadJSON(&om); err != nil {
				log.Error(err.Error())
				return
			}

			if om.Method != "subscription" {
				continue
			}

			switch om.Params.Data.OrderState {
			case "open":
				log.WithFields(log.Fields{
					"venue":    "deribit",
					"order":    om.Params.Data.OrderId,
					"quantity": om.Params.Data.FilledAmount,
				}).Debug("Fill")
			case "cancelled":
				log.WithFields(log.Fields{
					"venue": "deribit",
					"order": om.Params.Data.OrderId,
				}).Debug("Order cancelled")
				return
			case "filled":
				log.WithFields(log.Fields{
					"venue": "deribit",
					"order": om.Params.Data.OrderId,
				}).Info("Order filled")
				done <- true
				return
			}
		}
	}()

	return done
}

func (d *Deribit) canImprove(price, bestPrice float64, buy bool) bool {
	if buy {
		return price < bestPrice
	} else {
		return price > bestPrice
	}
}

func (d *Deribit) Trade(instrument string, contracts int, buy, reduce bool) {
	log.WithFields(log.Fields{
		"venue":      "deribit",
		"instrument": instrument,
		"contracts":  contracts,
		"buy":        buy,
		"reduce":     reduce,
	}).Info("Trade")

	var price, bp float64
	var orderId string
	var err error
	doneBestPrice := make(chan bool, 1)
	bestPrice := d.makeBestPrice(instrument, buy, doneBestPrice)
	done := d.orderStatus(instrument)
	ticker := time.NewTicker(10 * time.Millisecond)

	for {
		select {
		case <-done:
			doneBestPrice <- true
			return
		case <-ticker.C:
			if orderId == "" {
				price = bestPrice()
				orderId, err = d.PlaceOrder(instrument, contracts, price, buy, reduce)
				if err != nil {
					return
				}
			} else {
				bp = bestPrice()
				if d.canImprove(price, bp, buy) {
					price = bp
					err = d.EditOrder(orderId, contracts, price, reduce)
					if err != nil {
						return
					}
				}
			}
		}
	}
}
