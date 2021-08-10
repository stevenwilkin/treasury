package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type exposureMessage struct {
	Value float64 `json:"value"`
}

var exposureCmd = &cobra.Command{
	Use:   "exposure",
	Short: "Retrieve BTC long exposure",
	Run: func(cmd *cobra.Command, args []string) {
		var em exposureMessage
		get("/exposure", &em)

		fmt.Println(em.Value)
	},
}
