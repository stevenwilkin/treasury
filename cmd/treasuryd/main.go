package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/bybit"
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
	bybitExchange := &bybit.Bybit{
		ApiKey:    os.Getenv("BYBIT_API_KEY"),
		ApiSecret: os.Getenv("BYBIT_API_SECRET")}

	btcThbPrices := bitkubExchange.Price(symbol.BTCTHB)
	usdtThbPrices := bitkubExchange.Price(symbol.USDTTHB)
	deribitEquity := deribitExchange.Equity()
	bybitEquity := bybitExchange.Equity()

	go func() {
		for {
			select {
			case btcThb := <-btcThbPrices:
				statum.SetSymbol(symbol.BTCTHB, btcThb)
			case usdtThb := <-usdtThbPrices:
				statum.SetSymbol(symbol.USDTTHB, usdtThb)
			case deribitBtc := <-deribitEquity:
				statum.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
			case bybitBtc := <-bybitEquity:
				statum.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
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

func trapSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
	log.Info("Shutting down")
}

func main() {
	initLogger()
	initState()
	initPriceFeeds()
	initControlSocket()
	initWS()
	initWeb()
	trapSigInt()
}
