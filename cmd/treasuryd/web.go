// +build web

package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func initWS() {
	http.HandleFunc("/ws", serveWs)

	ch := statum.SubscribeToSymbols()

	go func() {
		for {
			sn := <-ch
			sendPrice(sn.Symbol, sn.Value)
		}
	}()
}

func initWeb() {
	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/", fs)

	go func() {
		log.Info("Listening on 0.0.0.0:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}