package deribit

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func (d *Deribit) bestBidAsk(instrument string, done chan bool) (*float64, *float64) {
	var bid, ask float64
	quoteChannel := fmt.Sprintf("quote.%s", instrument)

	c, err := d.subscribe([]string{quoteChannel})
	if err != nil {
		log.Error(err.Error())
		return &bid, &ask
	}

	go func() {
		var qm quoteMessage
		defer c.Close()

		for {
			select {
			case <-done:
				return
			default:
				if err = c.ReadJSON(&qm); err != nil {
					log.Error(err.Error())
					return
				}

				if qm.Method != "subscription" {
					continue
				}

				bid = qm.Params.Data.BestBidPrice
				ask = qm.Params.Data.BestAskPrice
			}
		}
	}()

	return &bid, &ask
}

func (d *Deribit) makeBestPrice(instrument string, buy bool, done chan bool) func() float64 {
	bid, ask := d.bestBidAsk(instrument, done)

	for *bid == 0 || *ask == 0 {
	}

	return func() float64 {
		if buy {
			return *bid
		} else {
			return *ask
		}
	}
}
