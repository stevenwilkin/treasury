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
	for s, p := range prices {
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
	for v, balances := range assets {
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

	if _, ok := assets[v]; !ok {
		assets[v] = map[asset.Asset]float64{}
	}

	assets[v][a] = q
}

func costHandler(w http.ResponseWriter, r *http.Request) {
	c, err := strconv.ParseFloat(r.FormValue("cost"), 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Cost - %f\n", c)

	cost = c
}

func pnlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pm := pnlMessage{
		Cost:          cost,
		Value:         totalValue(),
		Pnl:           pnl(),
		PnlPercentage: pnlPercentage(),
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
