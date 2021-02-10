package main

import (
	"fmt"

	"github.com/stevenwilkin/treasury/binance"
	"github.com/stevenwilkin/treasury/bitkub"
	"github.com/stevenwilkin/treasury/symbol"
	"github.com/stevenwilkin/treasury/xe"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	bitkub := &bitkub.BitKub{}
	binance := &binance.Binance{}
	xe := &xe.XE{}

	usdThb, _ := xe.GetPrice()
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
