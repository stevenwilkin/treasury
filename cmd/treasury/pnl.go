package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
		resp, err := client.Get("http://unix/pnl")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var pm pnlMessage

		json.Unmarshal(body, &pm)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Cost:  %f\n", pm.Cost)
		fmt.Printf("Value: %f\n", pm.Value)
		fmt.Printf("PnL:   %f\n", pm.Pnl)
		fmt.Printf("PnL %%: %.2f\n", pm.PnlPercentage)
	},
}
