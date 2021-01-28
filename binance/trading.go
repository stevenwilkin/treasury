package binance

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func (b *Binance) BestBidAsk(done chan bool) (*float64, *float64) {
	var bid, ask float64

	go func() {
		c, err := b.subscribe("btcusdt@bookTicker")
		if err != nil {
			return
		}
		defer c.Close()

		var bookTicker bookTickerMessage

		for {
			select {
			case <-done:
				return
			default:
				err := c.ReadJSON(&bookTicker)
				if err != nil {
					log.Error(err.Error())
					return
				}

				bid, _ = strconv.ParseFloat(bookTicker.BidPrice, 64)
				ask, _ = strconv.ParseFloat(bookTicker.AskPrice, 64)
			}
		}
	}()

	return &bid, &ask
}

func (b *Binance) makeBestPrice(buy bool, done chan bool) func() float64 {
	bid, ask := b.BestBidAsk(done)

	for *bid == 0 || *ask == 0 {
	}

	return func() float64 {
		if buy {
			return *bid
		} else {
			return *ask
		}
	}
}

func (b *Binance) orderStatus() (chan bool, chan float64) {
	done := make(chan bool, 1)
	fills := make(chan float64)
	var udm userDataMessage

	key, err := b.listenKey()
	if err != nil {
		log.Error(err.Error())
		return done, fills
	}

	c, err := b.subscribe(key)
	if err != nil {
		log.Error(err.Error())
		return done, fills
	}

	go func() {
		for {
			if err := c.ReadJSON(&udm); err != nil {
				log.Error(err.Error())
				continue
			}

			if udm.EventType != "executionReport" {
				continue
			}

			switch udm.OrderStatus {
			case "FILLED":
				done <- true
				log.WithFields(log.Fields{
					"venue":    "binance",
					"order_id": udm.OrderId,
				}).Info("Order filled")
				c.Close()
				return
			case "PARTIALLY_FILLED":
				fillQty, _ := strconv.ParseFloat(udm.FillQty, 64)
				fillPrice, _ := strconv.ParseFloat(udm.FillPrice, 64)
				cumFillQty, _ := strconv.ParseFloat(udm.CumFillQty, 64)
				fills <- fillQty
				log.WithFields(log.Fields{
					"venue":        "binance",
					"order_id":     udm.OrderId,
					"fill_qty":     fillQty,
					"fill_price":   fillPrice,
					"cum_fill_qty": cumFillQty,
				}).Debug("Fill")
			}
		}
	}()

	return done, fills
}

func (b *Binance) CancelAllOrders() {
	log.WithField("venue", "binance").Debug("Cancelling orders")
	v := url.Values{"symbol": {"BTCUSDT"}}
	b.doRequest("DELETE", "/api/v3/openOrders", v, true)
}

func (b *Binance) PlaceOrder(quantity, price float64, buy bool) int64 {
	log.WithFields(log.Fields{
		"venue":    "binance",
		"quantity": quantity,
		"price":    price,
		"buy":      buy,
	}).Debug("Placing order")

	side := "BUY"
	if !buy {
		side = "SELL"
	}

	v := url.Values{
		"symbol":   {"BTCUSDT"},
		"side":     {side},
		"type":     {"LIMIT_MAKER"},
		"quantity": {strconv.FormatFloat(quantity, 'f', 2, 64)},
		"price":    {strconv.FormatFloat(price, 'f', 2, 64)}}

	body, err := b.doRequest("POST", "/api/v3/order", v, true)
	if err != nil {
		return 0
	}

	var response createOrderResponse
	json.Unmarshal(body, &response)

	if response.OrderId == 0 {
		log.WithField("venue", "binance").Debug("Order rejected")
	} else {
		log.WithFields(log.Fields{
			"venue":    "binance",
			"order_id": response.OrderId,
		}).Debug("Order accepted")
	}

	return response.OrderId
}

func (b *Binance) enterOrder(q quantity, price float64, buy bool) {
	for q.remaining(price) > 0 && b.PlaceOrder(q.remaining(price), price, buy) == 0 {
	}
}

func (b *Binance) canImprove(price, bestPrice float64, buy bool) bool {
	if buy {
		return price < bestPrice
	} else {
		return price > bestPrice
	}
}

func (b *Binance) Trade(btc float64, buy bool) {
	log.WithFields(log.Fields{
		"venue":    "binance",
		"quantity": btc,
		"buy":      buy,
	}).Info("Trade")

	done, fills := b.orderStatus()
	doneBestPrice := make(chan bool, 1)
	bestPrice := b.makeBestPrice(buy, doneBestPrice)

	quantity := &btcQuantity{btc: btc}
	ticker := time.NewTicker(10 * time.Millisecond)
	price := bestPrice()
	b.enterOrder(quantity, price, buy)

	for {
		select {
		case <-done:
			quantity.done()
			doneBestPrice <- true
			return
		case fillQty := <-fills:
			quantity.fill(fillQty)
		case <-ticker.C:
			bp := bestPrice()
			if b.canImprove(price, bp, buy) {
				price = bp
				b.CancelAllOrders()
				b.enterOrder(quantity, price, buy)
			}
		}
	}
}
