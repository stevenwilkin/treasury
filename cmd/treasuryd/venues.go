package main

import (
	"os"

	"github.com/stevenwilkin/treasury/binance"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/bybit"
	"github.com/stevenwilkin/treasury/deribit"
	"github.com/stevenwilkin/treasury/ftx"
	"github.com/stevenwilkin/treasury/oanda"

	log "github.com/sirupsen/logrus"
)

type Venues struct {
	Binance *binance.Binance
	Bitkub  *bitkub.BitKub
	Deribit *deribit.Deribit
	Bybit   *bybit.Bybit
	Ftx     *ftx.FTX
	Oanda   *oanda.Oanda
}

func initVenues() {
	log.Info("Initialising venues")

	venues.Binance = &binance.Binance{}
	venues.Bitkub = &bitkub.BitKub{}
	venues.Deribit = &deribit.Deribit{
		ApiId:     os.Getenv("DERIBIT_API_ID"),
		ApiSecret: os.Getenv("DERIBIT_API_SECRET")}
	venues.Bybit = &bybit.Bybit{
		ApiKey:    os.Getenv("BYBIT_API_KEY"),
		ApiSecret: os.Getenv("BYBIT_API_SECRET")}
	venues.Ftx = &ftx.FTX{
		ApiKey:    os.Getenv("FTX_API_KEY"),
		ApiSecret: os.Getenv("FTX_API_SECRET")}
	venues.Oanda = &oanda.Oanda{
		AccountId: os.Getenv("OANDA_ACCOUNT_ID"),
		ApiKey:    os.Getenv("OANDA_API_KEY")}
}
