package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type fundingMessage struct {
	Current   float64 `json:"current"`
	Predicted float64 `json:"predicted"`
}

var fundingCmd = &cobra.Command{
	Use:   "funding",
	Short: "Retrieve funding",
	Run: func(cmd *cobra.Command, args []string) {
		var pm fundingMessage
		get("/funding", &pm)

		fmt.Printf("Current:   %f%%\n", pm.Current*100)
		fmt.Printf("Predicted: %f%%\n", pm.Predicted*100)
	},
}
