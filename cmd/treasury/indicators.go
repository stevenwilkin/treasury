package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type indicatorsMessage struct {
	THBPremium  float64 `json:"thb_premium"`
	USDTPremium float64 `json:"usdt_premium"`
}

var indicatorsCmd = &cobra.Command{
	Use:   "indicators",
	Short: "Retrieve indicators",
	Run: func(cmd *cobra.Command, args []string) {
		var im indicatorsMessage
		get("/indicators", &im)

		fmt.Printf("THB  Premium: %+.2f%%\n", im.THBPremium*100)
		fmt.Printf("USDT Premium: %+.2f%%\n", im.USDTPremium*100)
		fmt.Printf("Combined:     %+.2f%%\n", (im.THBPremium+im.USDTPremium)*100)
	},
}
