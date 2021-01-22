package zipmex

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Zipmex struct{}

type tickerMessage struct {
	O string `json:"o"`
}

func (z *Zipmex) subscribeToPrice(s symbol.Symbol) (*websocket.Conn, error) {
	var instrumentId int
	switch s {
	case symbol.BTCTHB:
		instrumentId = 26
	case symbol.USDTTHB:
		instrumentId = 53
	}

	u := url.URL{Scheme: "wss", Host: "api.exchange.zipmex.com", Path: "/WSGateway/"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	command := fmt.Sprintf(
		`{"m":0,"i":0,"n":"SubscribeTicker","o":"{\"OMSId\":1,\"InstrumentId\":%d,\"Interval\":900,\"IncludeLastCount\":1}"}`,
		instrumentId)
	if err = c.WriteMessage(websocket.TextMessage, []byte(command)); err != nil {
		return &websocket.Conn{}, err
	}

	return c, nil
}

func (z *Zipmex) Price(s symbol.Symbol) chan float64 {
	log.WithFields(log.Fields{
		"venue":  "zipmex",
		"symbol": s,
	}).Info("Subscribing to price")

	ch := make(chan float64)

	c, err := z.subscribeToPrice(s)
	if err != nil {
		return ch
	}

	go func() {
		defer c.Close()

		for {
			var ticker tickerMessage
			if err := c.ReadJSON(&ticker); err != nil {
				log.WithField("venue", "zipmex").Info("Reconnecting to price subscription")
				c.Close()
				c, err = z.subscribeToPrice(s)
				if err != nil {
					log.Error(err.Error())
					return
				}
				continue
			}

			var prices [][]float64
			json.Unmarshal([]byte(ticker.O), &prices)

			if len(prices) != 1 && len(prices[0]) < 5 {
				log.WithFields(log.Fields{
					"venue":  "zipmex",
					"symbol": s,
				}).Debug("Unexpected ticker message")
				continue
			}

			price := prices[0][4]
			log.WithFields(log.Fields{
				"venue":  "zipmex",
				"symbol": s,
				"value":  price,
			}).Debug("Received price")
			ch <- price
		}
	}()

	return ch
}
