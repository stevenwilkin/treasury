package main

import (
	"fmt"
	"os"

	"github.com/stevenwilkin/treasury/asset"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/ftx"
	"github.com/stevenwilkin/treasury/symbol"

	_ "github.com/joho/godotenv/autoload"
)

var (
	balances asset.Balances
	prices   symbol.Prices
)

func fetchData(ftx *ftx.FTX, bitkub *bitkub.BitKub) {
	for asset, balance := range ftx.Balances() {
		balances[asset] += balance
	}

	btcPrice := make(chan float64, 1)
	usdtPrice := make(chan float64, 1)

	go bitkub.Price(symbol.BTCTHB, btcPrice)
	go bitkub.Price(symbol.USDTTHB, usdtPrice)

	prices[symbol.BTCTHB] = <-btcPrice
	prices[symbol.USDTTHB] = <-usdtPrice
}

func displayResults() {
	fmt.Printf("BTC:     %f\n", balances[asset.BTC])
	fmt.Printf("USDT:    %f\n\n", balances[asset.USDT])

	fmt.Printf("BTCTHB:  %f\n", prices[symbol.BTCTHB])
	fmt.Printf("USDTTHB: %f\n\n", prices[symbol.USDTTHB])

	valueBTC := balances[asset.BTC] * prices[symbol.BTCTHB]
	valueUSDT := balances[asset.USDT] * prices[symbol.USDTTHB]
	total := valueBTC + valueUSDT
	fmt.Printf("Total:   %f\n", total)
}

func init() {
	balances = asset.Balances{
		asset.BTC:  0,
		asset.USDT: 0}

	balances[asset.USDT] += 201.71    // Blockfi
	balances[asset.USDT] += 106939.15 // Nexo

	prices = symbol.Prices{
		symbol.BTCTHB:  0,
		symbol.USDTTHB: 0}
}

func main() {
	bitkub := &bitkub.BitKub{}

	ftx := &ftx.FTX{
		ApiKey:    os.Getenv("FTX_API_KEY"),
		ApiSecret: os.Getenv("FTX_API_SECRET")}

	fetchData(ftx, bitkub)
	displayResults()
}
