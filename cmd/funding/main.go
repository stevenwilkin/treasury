package main

import (
	"fmt"

	"github.com/stevenwilkin/treasury/bybit"
)

func main() {
	b := &bybit.Bybit{}
	fundingRate, PredictedFundingRate := b.GetFundingRate()

	fmt.Printf("Funding:   %f%%\n", fundingRate*100)
	fmt.Printf("Predicted: %f%%\n", PredictedFundingRate*100)
}
