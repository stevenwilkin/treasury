// +build web

package main

import (
	"fmt"
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

	port := "8080"
	if wwwPort := os.Getenv("WWW_PORT"); len(wwwPort) > 0 {
		port = wwwPort
	}

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	go func() {
		log.Infof("Listening on 0.0.0.0:%s", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()
}
