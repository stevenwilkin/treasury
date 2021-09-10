package main

import (
	"fmt"

	"github.com/stevenwilkin/treasury/bybit"
)

func main() {
	b := &bybit.Bybit{}
	rates, _ := b.GetFundingRate()

	fmt.Printf("Funding:   %f%%\n", rates[0]*100)
	fmt.Printf("Predicted: %f%%\n", rates[1]*100)
}
