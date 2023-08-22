//go:build !noalerter

package daemon

import (
	"time"

	"github.com/stevenwilkin/treasury/alert"
	"github.com/stevenwilkin/treasury/telegram"

	log "github.com/sirupsen/logrus"
)

func (d *Daemon) initAlerter() {
	log.Info("Initialising alerter")

	notifier := telegram.NewFromEnv()

	d.alerter = alert.NewAlerter(d.state, notifier)
	d.alerter.Retrieve()

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			log.Debug("Checking alerts")
			d.alerter.CheckAlerts()
			d.alerter.Persist()
		}
	}()
}
