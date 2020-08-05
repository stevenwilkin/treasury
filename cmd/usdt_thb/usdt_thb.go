package main

import (
	"fmt"
	"os"

	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/oanda"
	"github.com/stevenwilkin/treasury/symbol"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	bitkub := &bitkub.BitKub{}
	oanda := &oanda.Oanda{
		AccountId: os.Getenv("OANDA_ACCOUNT_ID"),
		ApiKey:    os.Getenv("OANDA_API_KEY")}

	usdThb := oanda.Price(symbol.USDTHB)
	usdtThb := <-bitkub.Price(symbol.USDTTHB)
	difference := ((usdtThb - usdThb) / usdThb) * 100

	fmt.Printf("USDTTHB: %.2f\n", usdtThb)
	fmt.Printf("USDTHB:  %.2f\n", usdThb)
	fmt.Printf("Diff:    %.2f%%\n", difference)
}
