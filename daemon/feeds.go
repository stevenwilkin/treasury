package daemon

import (
	"reflect"
	"time"

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

func poll(feed interface{}) interface{} {
	returnType := reflect.ValueOf(feed).Type().Out(0)
	chType := reflect.ChanOf(reflect.BothDir, returnType)

	pollerF := func(_ []reflect.Value) []reflect.Value {
		ch := reflect.MakeChan(chType, 0)
		ticker := time.NewTicker(1 * time.Second)

		go func() {
			for {
				results := reflect.ValueOf(feed).Call([]reflect.Value{})
				if !results[1].IsNil() {
					ch.Close()
					return
				}

				ch.Send(results[0])
				<-ticker.C
			}
		}()

		return []reflect.Value{ch}
	}

	fnType := reflect.FuncOf([]reflect.Type{}, []reflect.Type{chType}, false)
	fn := reflect.MakeFunc(fnType, pollerF)

	return fn.Interface()
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
		poll(d.venues.Binance.GetBalances),
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
		poll(d.venues.XE.GetPrice),
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
		poll(d.venues.Bybit.GetEquityAndLeverage),
		func(equityAndLeverage [2]float64) {
			d.state.SetAsset(venue.Bybit, asset.BTC, equityAndLeverage[0])
			d.state.SetLeverageBybit(equityAndLeverage[1])
		})

	d.feedHandler.Add(
		feed.Funding,
		poll(d.venues.Bybit.GetFundingRate),
		func(funding [2]float64) {
			d.state.SetFundingRate(funding[0], funding[1])
		})

	d.feedHandler.Add(
		feed.LeverageDeribit,
		poll(d.venues.Deribit.GetLeverage),
		func(leverage float64) {
			d.state.SetLeverageDeribit(leverage)
		})
}
