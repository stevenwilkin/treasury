package main

import (
	"fmt"
	"os"

	"github.com/stevenwilkin/treasury/oanda"
	"github.com/stevenwilkin/treasury/symbol"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	o := &oanda.Oanda{
		AccountId: os.Getenv("OANDA_ACCOUNT_ID"),
		ApiKey:    os.Getenv("OANDA_API_KEY")}
	fmt.Println(<-o.Price(symbol.USDTHB))
}
