package main

import (
	"net/http"

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
