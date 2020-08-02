package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/venue"
)

func pricesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range statum.Symbols {
		pm.Prices[s.String()] = p
	}

	b, err := json.Marshal(pm)
	if err != nil {
		log.Println("error:", err)
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
		log.Println("error:", err)
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

	log.Printf("Set %s:%s to %f\n", v, a, q)

	statum.SetAsset(v, a, q)
}

func costHandler(w http.ResponseWriter, r *http.Request) {
	c, err := strconv.ParseFloat(r.FormValue("cost"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Cost - %f\n", c)

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
		log.Println("error:", err)
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

	return mux
}
