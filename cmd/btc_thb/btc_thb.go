package main

import (
	"fmt"
	"os"

	"github.com/stevenwilkin/treasury/binance"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/oanda"
	"github.com/stevenwilkin/treasury/symbol"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	bitkub := &bitkub.BitKub{}
	binance := &binance.Binance{}
	oanda := &oanda.Oanda{
		AccountId: os.Getenv("OANDA_ACCOUNT_ID"),
		ApiKey:    os.Getenv("OANDA_API_KEY")}

	btcThbPrices := make(chan float64, 1)
	btcUsdtPrices := make(chan float64, 1)

	go bitkub.Price(symbol.BTCTHB, btcThbPrices)
	go binance.Price(btcUsdtPrices)

	usdThb := oanda.Price(symbol.USDTHB)
	btcThb := <-btcThbPrices
	btcUsdt := <-btcUsdtPrices

	equivalent := btcThb / usdThb
	difference := equivalent - btcUsdt

	fmt.Printf("BTCTHB:  %.2f\n", btcThb)
	fmt.Printf("USDTHB:  %.2f\n", usdThb)
	fmt.Printf("Equiv:   %.2f\n", equivalent)
	fmt.Printf("BTCUSDT: %.2f\n", btcUsdt)
	fmt.Printf("Diff:    %.2f\n", difference)
}
