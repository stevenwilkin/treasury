package daemon

import (
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/handlers"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/venue"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Daemon struct {
	state       *state.State
	alerter     *alert.Alerter
	feedHandler *feed.Handler
	venues      venue.Venues
	conns       map[*websocket.Conn]bool
	m           sync.Mutex
}

const (
	socketPath = "/tmp/treasuryd.sock"
)

func (d *Daemon) initState() {
	log.Info("Initialising state")
	d.state = state.NewState()
	d.state.Load()

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			log.Debug("Persisting state")
			d.state.Save()
		}
	}()
}

func (d *Daemon) initVenues() {
	log.Info("Initialising venues")
	d.venues = venue.NewVenues()
}

func (d *Daemon) initControlSocket() {
	log.Info("Initialising control socket ", socketPath)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	err = os.Chmod(socketPath, 0777)
	if err != nil {
		log.Fatal("chmod error:", err)
	}

	h := handlers.NewHandler(d.state, d.alerter, d.feedHandler, d.venues)

	go func() {
		defer l.Close()
		log.Fatal(http.Serve(l, h.Mux()))
	}()
}

func (d *Daemon) Run() {
	d.initState()
	d.initAlerter()
	d.initVenues()
	d.initDataFeeds()
	d.initControlSocket()
	d.initWS()
}

func NewDaemon() *Daemon {
	return &Daemon{}
}
