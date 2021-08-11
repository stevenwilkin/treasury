package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type stateMessage struct {
	Assets          map[string]map[string]float64 `json:"assets"`
	Prices          map[string]float64            `json:"prices"`
	Exposure        float64                       `json:"exposure"`
	Cost            float64                       `json:"cost"`
	Value           float64                       `json:"value"`
	Pnl             float64                       `json:"pnl"`
	PnlPercentage   float64                       `json:"pnl_percentage"`
	LeverageDeribit float64                       `json:"leverage_deribit"`
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

	usdThb := state.Symbol(symbol.USDTHB)

	sm := stateMessage{
		Assets:          map[string]map[string]float64{},
		Prices:          map[string]float64{},
		Exposure:        state.Exposure(),
		Cost:            state.Cost / usdThb,
		Value:           state.TotalValue() / usdThb,
		Pnl:             state.Pnl() / usdThb,
		PnlPercentage:   state.PnlPercentage(),
		LeverageDeribit: state.GetLeverageDeribit()}

	for v, balances := range state.GetAssets() {
		sm.Assets[v.String()] = map[string]float64{}
		for a, q := range balances {
			sm.Assets[v.String()][a.String()] = q
		}
	}

	for s, p := range state.GetSymbols() {
		sm.Prices[s.String()] = p
	}

	if err := c.WriteJSON(sm); err != nil {
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
