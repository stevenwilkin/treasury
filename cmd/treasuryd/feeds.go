package main

import (
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	log "github.com/sirupsen/logrus"
)

func initDataFeeds() {
	log.Info("Initialising data feeds")
	feedHandler = feed.NewHandler()

	feedHandler.Add(venues.Binance.Price, func(btcUsdt float64) {
		statum.SetSymbol(symbol.BTCUSDT, btcUsdt)
	})

	feedHandler.AddArray(venues.Binance.Balances, func(balances [2]float64) {
		statum.SetAsset(venue.Binance, asset.BTC, balances[0])
		statum.SetAsset(venue.Binance, asset.USDT, balances[1])
	})

	feedHandler.Add(func() chan float64 {
		return venues.Bitkub.Price(symbol.BTCTHB)
	}, func(btcThb float64) {
		statum.SetSymbol(symbol.BTCTHB, btcThb)
	})

	feedHandler.Add(func() chan float64 {
		return venues.Bitkub.Price(symbol.USDTTHB)
	}, func(usdtThb float64) {
		statum.SetSymbol(symbol.USDTTHB, usdtThb)
	})

	feedHandler.Add(func() chan float64 {
		return venues.Oanda.Price(symbol.USDTHB)
	}, func(usdThb float64) {
		statum.SetSymbol(symbol.USDTHB, usdThb)
	})

	feedHandler.Add(venues.Deribit.Equity, func(deribitBtc float64) {
		statum.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
	})

	feedHandler.Add(venues.Bybit.Equity, func(bybitBtc float64) {
		statum.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
	})

	feedHandler.AddArray(venues.Bybit.FundingRate, func(funding [2]float64) {
		statum.SetFunding(funding[0], funding[1])
	})

	feedHandler.AddArray(venues.Ftx.Balances, func(balances [2]float64) {
		statum.SetAsset(venue.FTX, asset.BTC, balances[0])
		statum.SetAsset(venue.FTX, asset.USDT, balances[1])
	})
}
