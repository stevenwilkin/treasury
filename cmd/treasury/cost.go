package main

import (
	"net/url"

	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost [cost]",
	Short: "Set the total cost of the assets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		post("/cost", url.Values{"cost": {args[0]}})
	},
}
