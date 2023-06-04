package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type fundingMessage struct {
	Value float64 `json:"value"`
}

var fundingCmd = &cobra.Command{
	Use:   "funding",
	Short: "Retrieve funding",
	Run: func(cmd *cobra.Command, args []string) {
		var pm fundingMessage
		get("/funding", &pm)

		fmt.Printf("Funding: %f%%\n", pm.Value*100)
	},
}
