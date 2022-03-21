package daemon

import (
	"net/http"
	"os"

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
	LeverageBybit   float64                       `json:"leverage_bybit"`
}

type authMessage struct {
	Auth string `json:"auth"`
}

func (d *Daemon) sendState(c *websocket.Conn) error {
	log.Debug("Sending state")

	usdThb := d.state.Symbol(symbol.USDTHB)

	sm := stateMessage{
		Assets:          map[string]map[string]float64{},
		Prices:          map[string]float64{},
		Exposure:        d.state.Exposure(),
		Cost:            d.state.Cost / usdThb,
		Value:           d.state.TotalValue() / usdThb,
		Pnl:             d.state.Pnl() / usdThb,
		PnlPercentage:   d.state.PnlPercentage(),
		LeverageDeribit: d.state.GetLeverageDeribit(),
		LeverageBybit:   d.state.GetLeverageBybit()}

	for v, balances := range d.state.GetAssets() {
		sm.Assets[v.String()] = map[string]float64{}
		for a, q := range balances {
			sm.Assets[v.String()][a.String()] = q
		}
	}

	for s, p := range d.state.GetSymbols() {
		sm.Prices[s.String()] = p
	}

	if err := c.WriteJSON(sm); err != nil {
		return err
	}

	return nil
}

func (d *Daemon) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Debug("Accepting connection")

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}

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

	d.sendState(c)

	d.m.Lock()
	d.conns[c] = true
	d.m.Unlock()
}
