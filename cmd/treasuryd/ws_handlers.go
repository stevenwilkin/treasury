package main

import (
	"net/http"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type pricesMessage struct {
	Prices map[string]float64 `json:"prices"`
}

var (
	conns    = map[*websocket.Conn]bool{}
	upgrader = websocket.Upgrader{}
)

func sendState(c *websocket.Conn) {
	log.Debug("Sending initial state")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range statum.Symbols {
		pm.Prices[s.String()] = p
	}

	err := c.WriteJSON(pm)
	if err != nil {
		log.Error(err)
	}
}

func sendPrice(s symbol.Symbol, price float64) {
	log.WithFields(log.Fields{
		"symbol": s,
		"value":  price,
	}).Debug("Sending price to websockets")

	for c, _ := range conns {
		pm := pricesMessage{Prices: map[string]float64{s.String(): price}}

		err := c.WriteJSON(pm)
		if err != nil {
			log.Error(err)
			delete(conns, c)
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Debug("Accepting connection")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	conns[c] = true
	sendState(c)
}
