package bitkub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type BitKub struct{}

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

func (b *BitKub) PriceWS(s symbol.Symbol) chan float64 {
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

func (b *BitKub) GetPrice(s symbol.Symbol) (float64, error) {
	var tickerString string

	switch s {
	case symbol.BTCTHB:
		tickerString = "THB_BTC"
	case symbol.USDTTHB:
		tickerString = "THB_USDT"
	case symbol.USDCTHB:
		tickerString = "THB_USDC"
	}

	v := url.Values{"sym": {tickerString}}

	u := url.URL{
		Scheme:   "https",
		Host:     "api.bitkub.com",
		Path:     "/api/market/ticker",
		RawQuery: v.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response tickerResponse
	json.Unmarshal(body, &response)

	return response[tickerString].Last, nil
}

func (b *BitKub) Price(s symbol.Symbol) chan float64 {
	log.WithFields(log.Fields{
		"venue":  "bitkub",
		"symbol": s,
	}).Info("Polling price")

	ch := make(chan float64)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			price, err := b.GetPrice(s)
			if err != nil {
				log.WithField("venue", "bitkub").Warn(err.Error())
				close(ch)
				return
			}

			log.WithFields(log.Fields{
				"venue":  "bitkub",
				"symbol": s,
				"value":  price,
			}).Debug("Received price")

			ch <- price
			<-ticker.C
		}
	}()

	return ch
}
