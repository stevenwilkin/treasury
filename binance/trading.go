package binance

import (
	"encoding/json"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (b *Binance) BestBidAsk(done chan bool) (*float64, *float64) {
	var bid, ask float64

	go func() {
		c := b.subscribe("btcusdt@bookTicker")
		defer c.Close()

		var bookTicker bookTickerMessage

		for {
			select {
			case <-done:
				return
			default:
				err := c.ReadJSON(&bookTicker)
				if err != nil {
					log.Panic(err.Error())
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
		log.Panic(err.Error())
	}

	c := b.subscribe(key)

	go func() {
		for {
			if err := c.ReadJSON(&udm); err != nil {
				log.Panic(err.Error())
			}

			if udm.EventType != "executionReport" {
				continue
			}

			switch udm.OrderStatus {
			case "FILLED":
				log.WithFields(log.Fields{
					"venue":    "binance",
					"order_id": udm.OrderId,
				}).Info("Order filled")
				c.Close()
				done <- true
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

func (b *Binance) enterOrder(remaining *float64, price float64, buy bool) {
	for *remaining > 0 && b.PlaceOrder(*remaining, price, buy) == 0 {
	}
}

func (b *Binance) canImprove(price, bestPrice float64, buy bool) bool {
	if buy {
		return price < bestPrice
	} else {
		return price > bestPrice
	}
}

func (b *Binance) Trade(quantity float64, buy bool) {
	log.WithFields(log.Fields{
		"venue":    "binance",
		"quantity": quantity,
		"buy":      buy,
	}).Info("Trade")

	done, fills := b.orderStatus()
	doneBestPrice := make(chan bool)
	bestPrice := b.makeBestPrice(buy, doneBestPrice)

	remaining := quantity
	price := bestPrice()
	b.enterOrder(&remaining, price, buy)

	for {
		select {
		case <-done:
			remaining = 0
			doneBestPrice <- true
			return
		case fillQty := <-fills:
			remaining -= fillQty
		default:
			bp := bestPrice()
			if b.canImprove(price, bp, buy) {
				price = bp
				b.CancelAllOrders()
				b.enterOrder(&remaining, price, buy)
			}
		}
	}
}
