package bitkub

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type BitKub struct{}

type tickerMessage struct {
	Last float64
}

func (b *BitKub) subscribeToPrice(s symbol.Symbol) (*websocket.Conn, error) {
	var tickerString string

	switch s {
	case symbol.BTCTHB:
		tickerString = "thb_btc"
	case symbol.USDTTHB:
		tickerString = "thb_usdt"
	}

	path := fmt.Sprintf("websocket-api/market.ticker.%s", tickerString)
	u := url.URL{Scheme: "wss", Host: "api.bitkub.com", Path: path}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	return c, nil
}

func (b *BitKub) Price(s symbol.Symbol) chan float64 {
	log.WithFields(log.Fields{
		"venue":  "bitkub",
		"symbol": s,
	}).Info("Subscribing to price")

	ch := make(chan float64)

	c, err := b.subscribeToPrice(s)
	if err != nil {
		log.WithField("venue", "bitkub").Warn(err.Error())
		close(ch)
		return ch
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("venue", "bitkub").Warn(err.Error())
				c.Close()
				close(ch)
				return
			}

			var ticker tickerMessage
			json.Unmarshal(message, &ticker)

			log.WithFields(log.Fields{
				"venue":  "bitkub",
				"symbol": s,
				"value":  ticker.Last,
			}).Debug("Received price")
			ch <- ticker.Last
		}
	}()

	return ch
}
