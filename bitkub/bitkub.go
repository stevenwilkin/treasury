package bitkub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Bitkub struct{}

func symbolToTicker(s symbol.Symbol) string {
	var ticker string

	switch s {
	case symbol.BTCTHB:
		ticker = "THB_BTC"
	case symbol.USDTTHB:
		ticker = "THB_USDT"
	case symbol.USDCTHB:
		ticker = "THB_USDC"
	}

	return ticker
}
func (b *Bitkub) subscribeToPrice(s symbol.Symbol) (*websocket.Conn, error) {
	tickerString := symbolToTicker(s)

	path := fmt.Sprintf("websocket-api/market.ticker.%s", tickerString)
	u := url.URL{Scheme: "wss", Host: "api.bitkub.com", Path: path}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &websocket.Conn{}, err
	}

	return c, nil
}

func (b *Bitkub) PriceWS(s symbol.Symbol) chan float64 {
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

func (b *Bitkub) GetPrice(s symbol.Symbol) (float64, error) {
	var err error
	defer func() {
		if err != nil {
			log.WithField("venue", "bitkub").Warn(err.Error())
		}
	}()

	tickerString := symbolToTicker(s)

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
