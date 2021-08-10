package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type sizeMessage struct {
	Size int `json:"size"`
}

var sizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Retrieve size",
	Run: func(cmd *cobra.Command, args []string) {
		var pm sizeMessage
		get("/size", &pm)

		fmt.Println(pm.Size)
	},
}

var sizeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update size",
	Run: func(cmd *cobra.Command, args []string) {
		var pm sizeMessage
		get("/size/update", &pm)

		fmt.Println(pm.Size)
	},
}
