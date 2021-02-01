package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/stevenwilkin/treasury/binance"
	"github.com/stevenwilkin/treasury/bybit"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func main() {
	if level, err := log.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil {
		log.SetLevel(level)
	}

	spot := &binance.Binance{
		ApiKey:    os.Getenv("BINANCE_API_KEY"),
		ApiSecret: os.Getenv("BINANCE_API_SECRET")}
	perp := &bybit.Bybit{
		ApiKey:    os.Getenv("BYBIT_API_KEY"),
		ApiSecret: os.Getenv("BYBIT_API_SECRET")}

	var usd float64
	var rounds, i int64

	if len(os.Args) == 2 {
		usd, _ = strconv.ParseFloat(os.Args[1], 64)
		rounds = 1
	} else if len(os.Args) == 3 {
		usd, _ = strconv.ParseFloat(os.Args[1], 64)
		rounds, _ = strconv.ParseInt(os.Args[2], 10, 0)
	}

	if usd == 0 || rounds == 0 {
		fmt.Println("Invalid args")
		return
	}

	fmt.Printf("USD: %f Rounds: %d\n", usd, rounds)

	for i = 0; i < rounds; i++ {
		price := spot.TradeUSD(usd, true)
		perp.TradeWithLimit(int(usd), price, false, false)
	}
}
