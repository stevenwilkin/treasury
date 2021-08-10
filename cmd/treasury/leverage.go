package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type leverageMessage struct {
	Value float64 `json:"value"`
}

var leverageCmd = &cobra.Command{
	Use:   "leverage",
	Short: "Retrieve account leverage",
	Run: func(cmd *cobra.Command, args []string) {
		var em leverageMessage
		get("/leverage", &em)

		fmt.Printf("%.2f\n", em.Value)
	},
}
