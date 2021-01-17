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

	processFeed := func(chanF func() chan float64, processF func(float64)) {
		go func() {
			ch := chanF()
			for {
				processF(<-ch)
			}
		}()
	}

	processFeedArray := func(chanF func() chan [2]float64, processF func([2]float64)) {
		go func() {
			ch := chanF()
			for {
				processF(<-ch)
			}
		}()
	}

	processFeed(venues.Binance.Price, func(btcUsdt float64) {
		statum.SetSymbol(symbol.BTCUSDT, btcUsdt)
	})

	processFeedArray(venues.Binance.Balances, func(balances [2]float64) {
		statum.SetAsset(venue.Binance, asset.BTC, balances[0])
		statum.SetAsset(venue.Binance, asset.USDT, balances[1])
	})

	processFeed(func() chan float64 {
		return venues.Bitkub.Price(symbol.BTCTHB)
	}, func(btcThb float64) {
		statum.SetSymbol(symbol.BTCTHB, btcThb)
	})

	processFeed(func() chan float64 {
		return venues.Bitkub.Price(symbol.USDTTHB)
	}, func(usdtThb float64) {
		statum.SetSymbol(symbol.USDTTHB, usdtThb)
	})

	processFeed(func() chan float64 {
		return venues.Oanda.Price(symbol.USDTHB)
	}, func(usdThb float64) {
		statum.SetSymbol(symbol.USDTHB, usdThb)
	})

	processFeed(venues.Deribit.Equity, func(deribitBtc float64) {
		statum.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
	})

	processFeed(venues.Bybit.Equity, func(bybitBtc float64) {
		statum.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
	})

	processFeedArray(venues.Bybit.FundingRate, func(funding [2]float64) {
		statum.SetFunding(funding[0], funding[1])
	})

	processFeedArray(venues.Ftx.Balances, func(balances [2]float64) {
		statum.SetAsset(venue.FTX, asset.BTC, balances[0])
		statum.SetAsset(venue.FTX, asset.USDT, balances[1])
	})
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
		log.Fatal(err.Error())
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
