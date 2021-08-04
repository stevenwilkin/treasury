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
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/handlers"
	"github.com/stevenwilkin/treasury/state"
	"github.com/stevenwilkin/treasury/telegram"
	"github.com/stevenwilkin/treasury/venue"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

var (
	statum      *state.State
	alerter     *alert.Alerter
	feedHandler *feed.Handler
	venues      venue.Venues
)

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

	alerter = alert.NewAlerter(statum, notifier)
	alerter.Retrieve()

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			log.Debug("Checking alerts")
			alerter.CheckAlerts()
			alerter.Persist()
		}
	}()
}

func initVenues() {
	log.Info("Initialising venues")
	venues = venue.NewVenues()
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

	h := handlers.NewHandler(statum, alerter, feedHandler, venues)

	go func() {
		defer l.Close()
		log.Fatal(http.Serve(l, h.Mux()))
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
