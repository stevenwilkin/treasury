package main

import (
	"os"

	"github.com/stevenwilkin/treasury/ftx"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ftx := &ftx.FTX{
		ApiKey:    os.Getenv("FTX_API_KEY"),
		ApiSecret: os.Getenv("FTX_API_SECRET")}

	ftx.Trade(0.0001, false)
}
