// +build alerter

package main

import (
	"os"
	"strconv"
	"time"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/telegram"

	log "github.com/sirupsen/logrus"
)

func initAlerter() {
	log.Info("Initialising alerter")

	chatId, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		log.Fatal(err.Error())
	}

	notifier := &telegram.Telegram{
		ApiToken: os.Getenv("TELEGRAM_API_TOKEN"),
		ChatId:   chatId}

	alerter = alert.NewAlerter(state, notifier)
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
