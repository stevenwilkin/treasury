package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type pricesMessage struct {
	Prices map[string]float64 `json:"prices"`
}

type assetsMessage struct {
	Assets map[string]map[string]float64 `json:"assets"`
}

type pnlMessage struct {
	Cost          float64 `json:"cost"`
	Value         float64 `json:"value"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
}

const (
	socketPath = "/tmp/treasuryd.sock"
)

var (
	statum         *state.State
	bitkubExchange = &bitkub.BitKub{}
	conns          = map[*websocket.Conn]bool{}
	upgrader       = websocket.Upgrader{}
)

func sendState(c *websocket.Conn) {
	log.Debug("Sending initial state")

	pm := pricesMessage{Prices: map[string]float64{}}
	for s, p := range statum.Symbols {
		//log.Printf("%s - %f\n", s, p)
		pm.Prices[s.String()] = p
	}
	log.Printf("%v\n", pm)

	err := c.WriteJSON(pm)
	if err != nil {
		log.Error(err)
	}
}

func updatePrice(s symbol.Symbol, price float64) {
	log.Debugf("updatePrice - %s - %f\n", s, price)
	statum.SetSymbol(s, price)

	for c, _ := range conns {
		pm := pricesMessage{Prices: map[string]float64{s.String(): price}}

		err := c.WriteJSON(pm)
		if err != nil {
			log.Error(err)
			delete(conns, c)
		}
	}
}

func initPriceFeeds() {
	log.Info("Initialising price feeds")

	btcThbPrices := bitkubExchange.Price(symbol.BTCTHB)
	usdtThbPrices := bitkubExchange.Price(symbol.USDTTHB)

	go func() {
		for {
			select {
			case btcThb := <-btcThbPrices:
				updatePrice(symbol.BTCTHB, btcThb)
			case usdtThb := <-usdtThbPrices:
				updatePrice(symbol.USDTTHB, usdtThb)
			}
		}
	}()
}

func initState() {
	log.Info("Initialising state")
	statum = state.NewState()
	statum.Load()

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			log.Debug("Persisting state")
			statum.Save()
		}
	}()
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Debug("Accepting connection")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	//defer c.Close()
	conns[c] = true
	sendState(c)
}

func initWeb() {
	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", serveWs)
	log.Info("Listening on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initControlSocket() {
	log.Info("Initialising control socket", socketPath)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()

	mux := controlHandlers()
	log.Fatal(http.Serve(l, mux))
}

func initLogger() {
	if level, err := log.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil {
		log.SetLevel(level)
	}
}

func main() {
	initLogger()
	initState()
	initPriceFeeds()
	go initControlSocket()
	initWeb()
}
