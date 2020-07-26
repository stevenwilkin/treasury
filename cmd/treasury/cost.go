package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost [cost]",
	Short: "Set the total cost of the assets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.PostForm("http://unix/cost", url.Values{
			"cost": {args[0]}})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Failed")
			os.Exit(1)
		}
	},
}
