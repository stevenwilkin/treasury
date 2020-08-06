package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/deribit"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

var (
	statum *state.State
)

func initPriceFeeds() {
	log.Info("Initialising price feeds")

	bitkubExchange := &bitkub.BitKub{}
	deribitExchange := &deribit.Deribit{
		ApiId:     os.Getenv("DERIBIT_API_ID"),
		ApiSecret: os.Getenv("DERIBIT_API_SECRET")}

	btcThbPrices := bitkubExchange.Price(symbol.BTCTHB)
	usdtThbPrices := bitkubExchange.Price(symbol.USDTTHB)
	deribitEquity := deribitExchange.Equity()

	go func() {
		for {
			select {
			case btcThb := <-btcThbPrices:
				statum.SetSymbol(symbol.BTCTHB, btcThb)
			case usdtThb := <-usdtThbPrices:
				statum.SetSymbol(symbol.USDTTHB, usdtThb)
			case deribitBtc := <-deribitEquity:
				statum.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
			}
		}
	}()
}

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

func initWeb() {
	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/", fs)
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

	mux := controlHandlers()

	go func() {
		defer l.Close()
		log.Fatal(http.Serve(l, mux))
	}()
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
	initControlSocket()
	initWS()
	initWeb()
}
