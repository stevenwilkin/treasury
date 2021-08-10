package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type pnlMessage struct {
	Cost          float64 `json:"cost"`
	Value         float64 `json:"value"`
	Pnl           float64 `json:"pnl"`
	PnlPercentage float64 `json:"pnl_percentage"`
}

var pnlCmd = &cobra.Command{
	Use:   "pnl",
	Short: "Retrieve PnL",
	Run: func(cmd *cobra.Command, args []string) {
		var pm pnlMessage
		get("/pnl", &pm)

		fmt.Printf("Cost:  %f\n", pm.Cost)
		fmt.Printf("Value: %f\n", pm.Value)
		fmt.Printf("PnL:   %f\n", pm.Pnl)
		fmt.Printf("PnL %%: %.2f\n", pm.PnlPercentage)
	},
}

var pnlUsdCmd = &cobra.Command{
	Use:   "usd",
	Short: "Retrieve USD PnL",
	Run: func(cmd *cobra.Command, args []string) {
		var pm pnlMessage
		get("/pnl/usd", &pm)

		fmt.Printf("Cost:  %f\n", pm.Cost)
		fmt.Printf("Value: %f\n", pm.Value)
		fmt.Printf("PnL:   %f\n", pm.Pnl)
		fmt.Printf("PnL %%: %.2f\n", pm.PnlPercentage)
	},
}
