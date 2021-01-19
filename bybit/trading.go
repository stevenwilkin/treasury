package bybit

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (b *Bybit) orderRequest(params map[string]interface{}, path string) (string, error) {
	params["api_key"] = b.ApiKey
	params["timestamp"] = b.timestamp()
	params["sign"] = getSignature(params, b.ApiSecret)

	jsonRequest, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	u := url.URL{
		Scheme: "https",
		Host:   b.hostname(),
		Path:   path}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonRequest))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
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

	return response.Result.OrderId, nil
}

func (b *Bybit) PlaceOrder(amount int, price float64, buy, reduce bool) string {
	log.WithFields(log.Fields{
		"venue":  "bybit",
		"amount": amount,
		"price":  price,
	}).Info("Placing order")

	params := map[string]interface{}{
		"symbol":        "BTCUSD",
		"order_type":    "Limit",
		"qty":           strconv.Itoa(amount),
		"price":         strconv.FormatFloat(price, 'f', 2, 64),
		"time_in_force": "PostOnly"}

	if buy {
		params["side"] = "Buy"
	} else {
		params["side"] = "Sell"
	}

	if reduce {
		params["reduce_only"] = true
	}

	orderId, err := b.orderRequest(params, "/v2/private/order/create")
	if err != nil {
		log.Error(err.Error())
	}

	return orderId
}

func (b *Bybit) EditOrder(id string, price float64) string {
	log.WithFields(log.Fields{
		"venue": "bybit",
		"order": id,
		"price": price,
	}).Debug("Updating order")

	params := map[string]interface{}{
		"order_id":  id,
		"symbol":    "BTCUSD",
		"p_r_price": strconv.FormatFloat(price, 'f', 2, 64)}

	orderId, err := b.orderRequest(params, "/v2/private/order/replace")
	if err != nil {
		log.Error(err.Error())
	}

	return orderId
}

func highest(orders map[int64]float64) float64 {
	var result float64

	for _, x := range orders {
		if x > result {
			result = x
		}
	}

	return result
}

func lowest(orders map[int64]float64) float64 {
	var result float64

	for _, x := range orders {
		if result == 0.0 {
			result = x
		} else if x < result {
			result = x
		}
	}

	return result
}

func (b *Bybit) Trade(contracts int, buy, reduce bool) {
	var bestPrice, price float64
	var orderId string
	var pendingInitialOrder bool

	bids := map[int64]float64{}
	asks := map[int64]float64{}

	orderBookTopic := "orderBookL2_25.BTCUSD"
	orderTopic := "order"

	c, err := b.subscribe([]string{orderBookTopic, orderTopic})
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

		var response wsResponse
		json.Unmarshal(message, &response)

		switch response.Topic {
		case orderTopic:
			var orders []orderTopicData
			json.Unmarshal(response.Data, &orders)
			order := orders[0]

			switch order.OrderStatus {
			case "New":
				orderId = order.OrderId
			case "PartiallyFilled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        orderId,
					"cum_quantity": order.CumExecQty,
				}).Debug("Fill")
			case "Filled":
				log.WithFields(log.Fields{
					"venue": "bybit",
					"order": orderId,
				}).Info("Order filled")
				return
			case "Cancelled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        orderId,
					"quantity":     order.Qty,
					"cum_quantity": order.CumExecQty,
				}).Debug("Order cancelled")
				orderId = ""
				contracts = order.Qty - order.CumExecQty
				pendingInitialOrder = false
			default:
				pendingInitialOrder = false
			}
		case orderBookTopic:
			switch response.Type {
			case "snapshot":
				var snapshot snapshotData
				json.Unmarshal(response.Data, &snapshot)

				for _, order := range snapshot {
					p, _ := strconv.ParseFloat(order.Price, 64)

					if order.Side == "Buy" {
						bids[order.Id] = p
					} else {
						asks[order.Id] = p
					}
				}
			case "delta":
				var updates updateData
				json.Unmarshal(response.Data, &updates)

				for _, order := range updates.Delete {
					if order.Side == "Buy" {
						delete(bids, order.Id)
					} else {
						delete(asks, order.Id)
					}
				}

				for _, order := range updates.Insert {
					p, _ := strconv.ParseFloat(order.Price, 64)

					if order.Side == "Buy" {
						bids[order.Id] = p
					} else {
						asks[order.Id] = p
					}
				}
			}

			if buy {
				bestPrice = highest(bids)
			} else {
				bestPrice = lowest(asks)
			}

			if orderId == "" {
				if !pendingInitialOrder {
					price = bestPrice
					pendingInitialOrder = true
					b.PlaceOrder(contracts, price, buy, reduce)
				}
			} else if price != bestPrice {
				price = bestPrice
				b.EditOrder(orderId, price)
			}
		}
	}
}
