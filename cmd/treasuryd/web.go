// +build web

package main

import (
	"net/http"
	"os"
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
	path := "./www"
	if wwwRoot := os.Getenv("WWW_ROOT"); len(wwwRoot) > 0 {
		path = wwwRoot
	}

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	go func() {
		log.Info("Listening on 0.0.0.0:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}
