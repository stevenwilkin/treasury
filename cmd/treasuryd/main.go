package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/stevenwilkin/treasury/daemon"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

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

	d := daemon.NewDaemon()
	d.Run()

	trapSigInt()
}
