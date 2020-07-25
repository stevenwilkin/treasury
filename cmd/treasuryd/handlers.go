package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func controlHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/prices", pricesHandler)

	return mux
}
