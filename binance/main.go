package binance

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Binance struct{}

type tickerMessage struct {
	P string
}

func (b *Binance) Price() chan float64 {
	log.WithFields(log.Fields{
		"venue":  "binance",
		"symbol": "BTCUSDT",
	}).Info("Subscribing to price")

	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@aggTrade"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	ch := make(chan float64)

	go func() {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error(err)
				return
			}

			var ticker tickerMessage
			json.Unmarshal(message, &ticker)

			price, _ := strconv.ParseFloat(ticker.P, 64)
			ch <- price
		}
	}()

	return ch
}
