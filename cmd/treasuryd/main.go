package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/telegram"
	"github.com/stevenwilkin/treasury/venue"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

var (
	statum  *state.State
	alerter *alert.Alerter
	venues  Venues
)

func initDataFeeds() {
	log.Info("Initialising data feeds")

	btcUsdtPrices := venues.Binance.Price()
	btcThbPrices := venues.Bitkub.Price(symbol.BTCTHB)
	usdtThbPrices := venues.Bitkub.Price(symbol.USDTTHB)
	usdThbPrices := venues.Oanda.Price(symbol.USDTHB)
	deribitEquity := venues.Deribit.Equity()
	bybitEquity := venues.Bybit.Equity()
	bybitFundingRate := venues.Bybit.FundingRate()
	ftxBalances := venues.Ftx.Balances()

	go func() {
		for {
			select {
			case btcUsdt := <-btcUsdtPrices:
				statum.SetSymbol(symbol.BTCUSDT, btcUsdt)
			case btcThb := <-btcThbPrices:
				statum.SetSymbol(symbol.BTCTHB, btcThb)
			case usdtThb := <-usdtThbPrices:
				statum.SetSymbol(symbol.USDTTHB, usdtThb)
			case usdThb := <-usdThbPrices:
				statum.SetSymbol(symbol.USDTHB, usdThb)
			case deribitBtc := <-deribitEquity:
				statum.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
			case bybitBtc := <-bybitEquity:
				statum.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
			case funding := <-bybitFundingRate:
				statum.SetFunding(funding[0], funding[1])
			case balances := <-ftxBalances:
				statum.SetAsset(venue.FTX, asset.BTC, balances[0])
				statum.SetAsset(venue.FTX, asset.USDT, balances[1])
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

func initAlerter() {
	log.Info("Initialising alerter")

	chatId, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		log.Panic(err.Error())
	}

	notifier := &telegram.Telegram{
		ApiToken: os.Getenv("TELEGRAM_API_TOKEN"),
		ChatId:   chatId}

	alerter = alert.NewAlerter(notifier)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			log.Debug("Checking alerts")
			alerter.CheckAlerts()
		}
	}()
}

func initControlSocket() {
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
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c
	log.Info("Shutting down")
}

func main() {
	initLogger()
	initState()
	initAlerter()
	initVenues()
	initDataFeeds()
	initControlSocket()
	initWS()
	initWeb()
	trapSigInt()
}
