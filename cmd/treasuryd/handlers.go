package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	log "github.com/sirupsen/logrus"
)

type assetsMessage struct {
	Assets map[string]map[string]float64 `json:"assets"`
}

type pnlMessage struct {
	Cost          float64 `json:"cost"`
	Value         float64 `json:"value"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
}

type alertMessage struct {
	Active      bool   `json:"active"`
	Description string `json:"description"`
}

type fundingMessage struct {
	Current   float64 `float64:"current"`
	Predicted float64 `float64:"predicted"`
}

type feedsResponseItem struct {
	Active     bool
	LastUpdate time.Time
}
type feedsResponse struct {
	Feeds map[string]feedsResponseItem
}

func pricesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range statum.Symbols {
		pm.Prices[s.String()] = p
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	am := assetsMessage{Assets: map[string]map[string]float64{}}
	for v, balances := range statum.Assets {
		am.Assets[v.String()] = map[string]float64{}
		for a, q := range balances {
			am.Assets[v.String()][a.String()] = q
		}
	}

	b, err := json.Marshal(am)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	v, err := venue.FromString(r.FormValue("venue"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a, err := asset.FromString(r.FormValue("asset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q, err := strconv.ParseFloat(r.FormValue("quantity"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Set %s:%s to %f", v, a, q)

	statum.SetAsset(v, a, q)
}

func costHandler(w http.ResponseWriter, r *http.Request) {
	c, err := strconv.ParseFloat(r.FormValue("cost"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Cost - %f", c)

	statum.SetCost(c)
}

func pnlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pnlMessage{
		Cost:          statum.Cost,
		Value:         statum.TotalValue(),
		Pnl:           statum.Pnl(),
		PnlPercentage: statum.PnlPercentage(),
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func pnlUsdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	usdThb := statum.Symbol(symbol.USDTHB)

	pm := pnlMessage{
		Cost:          statum.Cost / usdThb,
		Value:         statum.TotalValue() / usdThb,
		Pnl:           statum.Pnl() / usdThb,
		PnlPercentage: statum.PnlPercentage(),
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func alertsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := alerter.Alerts()
	am := make([]alertMessage, len(alerts))

	for i, alert := range alerts {
		am[i] = alertMessage{
			Active:      alert.Active(),
			Description: alert.Description()}
	}

	b, err := json.Marshal(am)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func clearAlertsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Clearing alerts")
	alerter.ClearAlerts()
}

func priceAlertsHandler(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.ParseFloat(r.FormValue("value"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Setting price alert - %f", v)

	a := alert.NewPriceAlert(statum, symbol.BTCUSDT, v)
	alerter.AddAlert(a)
}

func fundingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	current, predicted := statum.Funding()

	fm := fundingMessage{
		Current:   current,
		Predicted: predicted,
	}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func fundingAlertsHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Setting funding alert")

	a := alert.NewFundingAlert(statum)
	alerter.AddAlert(a)
}

func exposureHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Value float64 `json:"value"`
	}{
		Value: statum.Exposure()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func sizeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Size int `json:"size"`
	}{
		Size: statum.Size()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func updateSizeHandler(w http.ResponseWriter, r *http.Request) {
	size := venues.Deribit.GetSize() + venues.Bybit.GetSize()
	log.Infof("Setting size to %d", size)

	statum.SetSize(size)

	sizeHandler(w, r)
}

func feedsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fr := feedsResponse{Feeds: map[string]feedsResponseItem{}}

	for feed, status := range feedHandler.Status() {
		fr.Feeds[feed.String()] = feedsResponseItem{
			Active: status.Active, LastUpdate: status.LastUpdate}
	}

	b, err := json.Marshal(fr)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func feedsReactivateHandler(w http.ResponseWriter, r *http.Request) {
	f, err := feed.FromString(r.FormValue("feed"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feedHandler.Reactivate(f)
}

func indicatorsHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]float64{
		"thb_premium":  statum.THBPremium(),
		"usdt_premium": statum.USDTPremium()}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func controlHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/prices", pricesHandler)
	mux.HandleFunc("/assets", assetsHandler)
	mux.HandleFunc("/set", setHandler)
	mux.HandleFunc("/cost", costHandler)
	mux.HandleFunc("/pnl", pnlHandler)
	mux.HandleFunc("/pnl/usd", pnlUsdHandler)
	mux.HandleFunc("/alerts", alertsHandler)
	mux.HandleFunc("/alerts/clear", clearAlertsHandler)
	mux.HandleFunc("/alerts/price", priceAlertsHandler)
	mux.HandleFunc("/alerts/funding", fundingAlertsHandler)
	mux.HandleFunc("/funding", fundingHandler)
	mux.HandleFunc("/exposure", exposureHandler)
	mux.HandleFunc("/size", sizeHandler)
	mux.HandleFunc("/size/update", updateSizeHandler)
	mux.HandleFunc("/feeds", feedsHandler)
	mux.HandleFunc("/feeds/reactivate", feedsReactivateHandler)
	mux.HandleFunc("/indicators", indicatorsHandler)

	return mux
}
