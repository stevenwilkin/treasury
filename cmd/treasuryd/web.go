// +build web

package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func initWS() {
	http.HandleFunc("/ws", serveWs)

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			for c, _ := range conns {
				if err := sendState(c); err != nil {
					log.Debug(err)
					m.Lock()
					delete(conns, c)
					m.Unlock()
				}
			}

			<-ticker.C
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
