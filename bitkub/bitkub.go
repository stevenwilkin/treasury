package bitkub

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
)

type BitKub struct{}

type tickerMessage struct {
	Last float64
}

func (b *BitKub) Price(s symbol.Symbol, price chan<- float64) {
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
		panic(err.Error())
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			break
		}
		//fmt.Printf("recv: %s\n", message)

		var ticker tickerMessage
		json.Unmarshal(message, &ticker)

		//fmt.Println(ticker.Last)
		price <- ticker.Last
	}
}
