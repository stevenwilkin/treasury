package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	s *state.State
	a *alert.Alerter
	f *feed.Handler
	v venue.Venues
}

func NewHandler(s *state.State, a *alert.Alerter, f *feed.Handler, v venue.Venues) *Handler {
	return &Handler{
		a: a,
		s: s,
		f: f,
		v: v}
}

func (h *Handler) Prices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range h.s.GetSymbols() {
		pm.Prices[s.String()] = p
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) Assets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	am := assetsMessage{Assets: map[string]map[string]float64{}}
	for v, balances := range h.s.GetAssets() {
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

func (h *Handler) SetAsset(w http.ResponseWriter, r *http.Request) {
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

	h.s.SetAsset(v, a, q)
}

func (h *Handler) SetCost(w http.ResponseWriter, r *http.Request) {
	c, err := strconv.ParseFloat(r.FormValue("cost"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Cost - %f", c)

	h.s.SetCost(c)
}

func (h *Handler) PnL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pnlMessage{
		Cost:          h.s.Cost,
		Value:         h.s.TotalValue(),
		Pnl:           h.s.Pnl(),
		PnlPercentage: h.s.PnlPercentage(),
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) PnLUSD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	usdThb := h.s.Symbol(symbol.USDTHB)

	pm := pnlMessage{
		Cost:          h.s.Cost / usdThb,
		Value:         h.s.TotalValue() / usdThb,
		Pnl:           h.s.Pnl() / usdThb,
		PnlPercentage: h.s.PnlPercentage(),
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) Alerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := h.a.Alerts()
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

func (h *Handler) ClearAlerts(w http.ResponseWriter, r *http.Request) {
	log.Info("Clearing alerts")
	h.a.ClearAlerts()
}

func (h *Handler) AddPriceAlert(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.ParseFloat(r.FormValue("value"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Setting price alert - %f", v)

	h.a.AddPriceAlert(v)
}

func (h *Handler) Funding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	funding := h.s.GetFundingRate()

	fm := fundingMessage{Value: funding}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) AddFundingAlert(w http.ResponseWriter, r *http.Request) {
	log.Infof("Setting funding alert")

	h.a.AddFundingAlert()
}

func (h *Handler) Leverage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Deribit float64 `json:"deribit"`
		Bybit   float64 `json:"bybit"`
	}{
		Deribit: h.s.GetLeverageDeribit(),
		Bybit:   h.s.GetLeverageBybit()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) AddLeverageAlert(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.ParseFloat(r.FormValue("value"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Setting leverage alert - %f", v)

	h.a.AddLeverageAlert(v)
}

func (h *Handler) Exposure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Value float64 `json:"value"`
	}{
		Value: h.s.Exposure()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) Size(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Size int `json:"size"`
	}{
		Size: h.s.GetSize()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) UpdateSize(w http.ResponseWriter, r *http.Request) {
	size := h.v.Deribit.GetSize() + h.v.Bybit.GetSize()
	log.Infof("Setting size to %d", size)

	h.s.SetSize(size)

	h.Size(w, r)
}

func (h *Handler) Feeds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fr := feedsResponse{Feeds: map[string]feedsResponseItem{}}

	for feed, status := range h.f.Status() {
		fr.Feeds[feed.String()] = feedsResponseItem{
			Active: status.Active, LastUpdate: status.LastUpdate}
	}

	b, err := json.Marshal(fr)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) ReactivateFeed(w http.ResponseWriter, r *http.Request) {
	f, err := feed.FromString(r.FormValue("feed"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.f.Reactivate(f)
}

func (h *Handler) Indicators(w http.ResponseWriter, r *http.Request) {
	data := map[string]float64{
		"thb_premium":  h.s.THBPremium(),
		"usdt_premium": h.s.USDTPremium()}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) Loan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fm := struct {
		Loan float64 `json:"loan"`
	}{
		Loan: h.s.GetLoan()}

	b, err := json.Marshal(fm)
	if err != nil {
		log.Error(err)
	}

	w.Write(b)
}

func (h *Handler) SetLoan(w http.ResponseWriter, r *http.Request) {
	l, err := strconv.ParseFloat(r.FormValue("loan"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Loan - %f", l)

	h.s.SetLoan(l)
}

func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/prices", h.Prices)
	mux.HandleFunc("/assets", h.Assets)
	mux.HandleFunc("/set", h.SetAsset)
	mux.HandleFunc("/cost", h.SetCost)
	mux.HandleFunc("/pnl", h.PnL)
	mux.HandleFunc("/pnl/usd", h.PnLUSD)
	mux.HandleFunc("/alerts", h.Alerts)
	mux.HandleFunc("/alerts/clear", h.ClearAlerts)
	mux.HandleFunc("/alerts/price", h.AddPriceAlert)
	mux.HandleFunc("/alerts/funding", h.AddFundingAlert)
	mux.HandleFunc("/alerts/leverage", h.AddLeverageAlert)
	mux.HandleFunc("/funding", h.Funding)
	mux.HandleFunc("/exposure", h.Exposure)
	mux.HandleFunc("/leverage", h.Leverage)
	mux.HandleFunc("/size", h.Size)
	mux.HandleFunc("/size/update", h.UpdateSize)
	mux.HandleFunc("/feeds", h.Feeds)
	mux.HandleFunc("/feeds/reactivate", h.ReactivateFeed)
	mux.HandleFunc("/indicators", h.Indicators)
	mux.HandleFunc("/loan", h.Loan)
	mux.HandleFunc("/loan/set", h.SetLoan)

	return mux
}
