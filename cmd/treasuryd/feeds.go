package main

import (
	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/feed"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/venue"

	log "github.com/sirupsen/logrus"
)

func curry(f func(s symbol.Symbol) chan float64, s symbol.Symbol) func() chan float64 {
	return func() chan float64 {
		return f(s)
	}
}

func initDataFeeds() {
	log.Info("Initialising data feeds")
	feedHandler = feed.NewHandler()

	feedHandler.Add(
		feed.BTCUSDT,
		venues.Binance.Price,
		func(btcUsdt float64) {
			state.SetSymbol(symbol.BTCUSDT, btcUsdt)
		})

	feedHandler.Add(
		feed.Binance,
		venues.Binance.Balances,
		func(balances [3]float64) {
			state.SetAsset(venue.Binance, asset.BTC, balances[0])
			state.SetAsset(venue.Binance, asset.USDT, balances[1])
			state.SetAsset(venue.Binance, asset.USDC, balances[2])
		})

	feedHandler.Add(
		feed.BTCTHB,
		curry(venues.Bitkub.PriceWS, symbol.BTCTHB),
		func(btcThb float64) {
			state.SetSymbol(symbol.BTCTHB, btcThb)
		})

	feedHandler.Add(
		feed.USDTTHB,
		curry(venues.Bitkub.PriceWS, symbol.USDTTHB),
		func(usdtThb float64) {
			state.SetSymbol(symbol.USDTTHB, usdtThb)
		})

	feedHandler.Add(
		feed.USDCTHB,
		curry(venues.Bitkub.PriceWS, symbol.USDCTHB),
		func(usdcThb float64) {
			state.SetSymbol(symbol.USDCTHB, usdcThb)
		})

	feedHandler.Add(
		feed.USDTHB,
		venues.XE.Price,
		func(usdThb float64) {
			state.SetSymbol(symbol.USDTHB, usdThb)
		})

	feedHandler.Add(
		feed.Deribit,
		venues.Deribit.Equity,
		func(deribitBtc float64) {
			state.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
		})

	feedHandler.Add(
		feed.Bybit,
		venues.Bybit.Equity,
		func(bybitBtc float64) {
			state.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
		})

	feedHandler.Add(
		feed.Funding,
		venues.Bybit.FundingRate,
		func(funding [2]float64) {
			state.SetFunding(funding[0], funding[1])
		})

	feedHandler.Add(
		feed.FTX,
		venues.Ftx.Balances,
		func(balances [2]float64) {
			state.SetAsset(venue.FTX, asset.BTC, balances[0])
			state.SetAsset(venue.FTX, asset.USDT, balances[1])
		})

	feedHandler.Add(
		feed.LeverageDeribit,
		venues.Deribit.Leverage,
		func(leverage float64) {
			state.SetLeverageDeribit(leverage)
		})

	feedHandler.Add(
		feed.LeverageBybit,
		venues.Bybit.Leverage,
		func(leverage float64) {
			state.SetLeverageBybit(leverage)
		})
}
