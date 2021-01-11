package binance

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Binance struct{}

func (b *Binance) subscribeToPrice() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@aggTrade"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	return c
}

func (b *Binance) Price() chan float64 {
	log.WithFields(log.Fields{
		"venue":  "binance",
		"symbol": "BTCUSDT",
	}).Info("Subscribing to price")

	c := b.subscribeToPrice()
	ch := make(chan float64)

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "binance").Info("Reconnecting to price subscription")
				c.Close()
				c = b.subscribeToPrice()
				continue
			}

			var ticker tickerMessage
			json.Unmarshal(message, &ticker)

			price, _ := strconv.ParseFloat(ticker.P, 64)
			ch <- price
		}
	}()

	return ch
}
