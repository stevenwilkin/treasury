package daemon

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func (d *Daemon) initWS() {
	d.conns = map[*websocket.Conn]bool{}

	http.HandleFunc("/ws", d.serveWs)

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			for c, _ := range d.conns {
				if err := d.sendState(c); err != nil {
					log.Debug(err)
					d.m.Lock()
					delete(d.conns, c)
					d.m.Unlock()
				}
			}

			<-ticker.C
		}
	}()

	port := "8080"
	if wwwPort := os.Getenv("WWW_PORT"); len(wwwPort) > 0 {
		port = wwwPort
	}

	go func() {
		log.Infof("Listening on 0.0.0.0:%s", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()
}
