package bybit

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (b *Bybit) PlaceOrder(contracts int, price float64, buy, reduce bool) string {
	log.WithFields(log.Fields{
		"venue":     "bybit",
		"contracts": contracts,
		"price":     price,
		"buy":       buy,
	}).Debug("Placing order")

	params := map[string]interface{}{
		"symbol":        "BTCUSD",
		"order_type":    "Limit",
		"qty":           strconv.Itoa(contracts),
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

func (b *Bybit) orderStatus() (chan bool, chan int) {
	done := make(chan bool)
	fillsOnCancel := make(chan int)
	orderTopic := "order"

	c, err := b.subscribe([]string{orderTopic})
	if err != nil {
		log.Error(err.Error())
		return done, fillsOnCancel
	}

	go func() {
		defer c.Close()
		var orders orderTopicData

		for {
			if err := c.ReadJSON(&orders); err != nil {
				log.Error(err)
				return
			}

			if orders.Topic != orderTopic {
				continue
			}

			order := orders.Data[0]

			switch order.OrderStatus {
			case "PartiallyFilled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        order.OrderId,
					"price":        order.Price,
					"cum_quantity": order.CumExecQty,
				}).Debug("Fill")
			case "Filled":
				log.WithFields(log.Fields{
					"venue":    "bybit",
					"order":    order.OrderId,
					"price":    order.Price,
					"quantity": order.Qty,
				}).Info("Order filled")
				done <- true
				return
			case "Cancelled":
				log.WithFields(log.Fields{
					"venue":        "bybit",
					"order":        order.OrderId,
					"price":        order.Price,
					"quantity":     order.Qty,
					"cum_quantity": order.CumExecQty,
				}).Debug("Order cancelled")
				fillsOnCancel <- order.CumExecQty
			}
		}
	}()

	return done, fillsOnCancel
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

func (b *Bybit) bestBidAsk(done chan bool) (*float64, *float64) {
	var bid, ask float64
	bids := map[int64]float64{}
	asks := map[int64]float64{}

	orderBookTopic := "orderBookL2_25.BTCUSD"

	c, err := b.subscribe([]string{orderBookTopic})
	if err != nil {
		log.Error(err.Error())
		return &bid, &ask
	}

	go func() {
		defer c.Close()
		var response wsResponse

		for {
			select {
			case <-done:
				return
			default:
				if err := c.ReadJSON(&response); err != nil {
					log.Error(err)
					return
				}

				if response.Topic != orderBookTopic {
					continue
				}

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

				bid = highest(bids)
				ask = lowest(asks)
			}
		}
	}()

	return &bid, &ask
}

func (b *Bybit) makeBestPrice(buy bool, limit float64, done chan bool) func() float64 {
	bid, ask := b.bestBidAsk(done)

	return func() float64 {
		for *bid == 0 || *ask == 0 {
		}

		if buy {
			if limit == 0 {
				return *bid
			} else {
				return math.Min(*bid, limit)
			}
		} else {
			if limit == 0 {
				return *ask
			} else {
				return math.Max(*ask, limit)
			}
		}
	}
}

func (b *Bybit) canImprove(price, bestPrice float64, buy bool) bool {
	if buy {
		return price < bestPrice
	} else {
		return price > bestPrice
	}
}

func (b *Bybit) TradeWithLimit(contracts int, limit float64, buy, reduce bool) {
	log.WithFields(log.Fields{
		"venue":     "bybit",
		"contracts": contracts,
		"limit":     limit,
		"buy":       buy,
		"reduce":    reduce,
	}).Info("Trade")

	var orderId string
	var price, bp float64
	var fillQty int
	remaining := contracts

	doneBestPrice := make(chan bool)
	bestPrice := b.makeBestPrice(buy, limit, doneBestPrice)
	done, fillsOnCancel := b.orderStatus()
	ticker := time.NewTicker(10 * time.Millisecond)

	for {
		select {
		case <-done:
			remaining = 0
			doneBestPrice <- true
			return
		case fillQty = <-fillsOnCancel:
			remaining -= fillQty
			orderId = ""
		case <-ticker.C:
			if remaining == 0 {
				continue
			} else if orderId == "" {
				price = bestPrice()
				orderId = b.PlaceOrder(remaining, price, buy, reduce)
			} else {
				bp = bestPrice()
				if b.canImprove(price, bp, buy) {
					price = bp
					b.EditOrder(orderId, bp)
				}
			}
		}
	}
}

func (b *Bybit) Trade(contracts int, buy, reduce bool) {
	b.TradeWithLimit(contracts, 0, buy, reduce)
}
