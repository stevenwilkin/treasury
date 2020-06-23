package binance

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

type Binance struct{}

type tickerMessage struct {
	P string
}

func (b *Binance) Price(prices chan<- float64) {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@aggTrade"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err.Error())
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		//fmt.Printf("recv: %s\n", message)

		var ticker tickerMessage
		json.Unmarshal(message, &ticker)

		price, _ := strconv.ParseFloat(ticker.P, 64)
		//fmt.Println(price)
		prices <- price
	}
}
