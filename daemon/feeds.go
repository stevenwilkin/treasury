package daemon

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

func (d *Daemon) initDataFeeds() {
	log.Info("Initialising data feeds")
	d.feedHandler = feed.NewHandler()

	d.feedHandler.Add(
		feed.BTCUSDT,
		d.venues.Binance.Price,
		func(btcUsdt float64) {
			d.state.SetSymbol(symbol.BTCUSDT, btcUsdt)
		})

	d.feedHandler.Add(
		feed.Binance,
		d.venues.Binance.Balances,
		func(balances [3]float64) {
			d.state.SetAsset(venue.Binance, asset.BTC, balances[0])
			d.state.SetAsset(venue.Binance, asset.USDT, balances[1])
			d.state.SetAsset(venue.Binance, asset.USDC, balances[2])
		})

	d.feedHandler.Add(
		feed.BTCTHB,
		curry(d.venues.Bitkub.PriceWS, symbol.BTCTHB),
		func(btcThb float64) {
			d.state.SetSymbol(symbol.BTCTHB, btcThb)
		})

	d.feedHandler.Add(
		feed.USDTTHB,
		curry(d.venues.Bitkub.PriceWS, symbol.USDTTHB),
		func(usdtThb float64) {
			d.state.SetSymbol(symbol.USDTTHB, usdtThb)
		})

	d.feedHandler.Add(
		feed.USDTHB,
		d.venues.XE.Price,
		func(usdThb float64) {
			d.state.SetSymbol(symbol.USDTHB, usdThb)
		})

	d.feedHandler.Add(
		feed.Deribit,
		d.venues.Deribit.Equity,
		func(deribitBtc float64) {
			d.state.SetAsset(venue.Deribit, asset.BTC, deribitBtc)
		})

	d.feedHandler.Add(
		feed.Bybit,
		d.venues.Bybit.Equity,
		func(bybitBtc float64) {
			d.state.SetAsset(venue.Bybit, asset.BTC, bybitBtc)
		})

	d.feedHandler.Add(
		feed.Funding,
		d.venues.Bybit.FundingRate,
		func(funding [2]float64) {
			d.state.SetFunding(funding[0], funding[1])
		})

	d.feedHandler.Add(
		feed.FTX,
		d.venues.Ftx.Balances,
		func(balances [2]float64) {
			d.state.SetAsset(venue.FTX, asset.BTC, balances[0])
			d.state.SetAsset(venue.FTX, asset.USDT, balances[1])
		})

	d.feedHandler.Add(
		feed.LeverageDeribit,
		d.venues.Deribit.Leverage,
		func(leverage float64) {
			d.state.SetLeverageDeribit(leverage)
		})

	d.feedHandler.Add(
		feed.LeverageBybit,
		d.venues.Bybit.Leverage,
		func(leverage float64) {
			d.state.SetLeverageBybit(leverage)
		})
}
