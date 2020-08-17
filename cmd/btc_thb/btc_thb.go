package main

import (
	"fmt"
	"os"

	"github.com/stevenwilkin/treasury/binance"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/oanda"
	"github.com/stevenwilkin/treasury/symbol"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	bitkub := &bitkub.BitKub{}
	binance := &binance.Binance{}
	oanda := &oanda.Oanda{
		AccountId: os.Getenv("OANDA_ACCOUNT_ID"),
		ApiKey:    os.Getenv("OANDA_API_KEY")}

	usdThb := oanda.GetPrice(symbol.USDTHB)
	btcThb := <-bitkub.Price(symbol.BTCTHB)
	btcUsdt := <-binance.Price()

	equivalent := btcThb / usdThb
	difference := equivalent - btcUsdt
	percentage := (difference / btcUsdt) * 100

	fmt.Printf("BTCTHB:  %.2f\n", btcThb)
	fmt.Printf("USDTHB:  %.2f\n", usdThb)
	fmt.Printf("Equiv:   %.2f\n", equivalent)
	fmt.Printf("BTCUSDT: %.2f\n", btcUsdt)
	fmt.Printf("Diff:    %.2f\n", difference)
	fmt.Printf("         %.2f%%\n", percentage)
}
