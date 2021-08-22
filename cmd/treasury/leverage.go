package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type leverageMessage struct {
	Deribit float64 `json:"deribit"`
	Bybit   float64 `json:"bybit"`
}

var leverageCmd = &cobra.Command{
	Use:   "leverage",
	Short: "Retrieve account leverage",
	Run: func(cmd *cobra.Command, args []string) {
		var em leverageMessage
		get("/leverage", &em)

		if em.Deribit > 0 {
			fmt.Printf("Deribit: %.2f\n", em.Deribit)
		}

		if em.Bybit > 0 {
			fmt.Printf("Bybit:   %.2f\n", em.Bybit)
		}
	},
}
