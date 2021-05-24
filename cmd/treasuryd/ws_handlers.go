package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type pricesMessage struct {
	Prices map[string]float64 `json:"prices"`
}

type authMessage struct {
	Auth string `json:"auth"`
}

var (
	conns    = map[*websocket.Conn]bool{}
	upgrader = websocket.Upgrader{}
	m        = sync.Mutex{}
)

func sendState(c *websocket.Conn) error {
	log.Debug("Sending state")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range statum.Symbols {
		pm.Prices[s.String()] = p
	}

	if err := c.WriteJSON(pm); err != nil {
		return err
	}

	return nil
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Debug("Accepting connection")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn(err)
		return
	}

	if authToken := os.Getenv("WS_AUTH_TOKEN"); len(authToken) > 0 {
		var am authMessage
		if err := c.ReadJSON(&am); err != nil {
			log.Warn(err)
			return
		}

		if am.Auth != authToken {
			log.Info("Unauthenticated")
			c.WriteMessage(websocket.TextMessage, []byte(`{"error":"unauthenticated"}`))
			c.Close()
			return
		}
	}

	m.Lock()
	conns[c] = true
	m.Unlock()

	sendState(c)
}
