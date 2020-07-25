package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [venue] [asset] [quantity]",
	Short: "Set quantity of an asset within a venue",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.PostForm("http://unix/set", url.Values{
			"venue":    {args[0]},
			"asset":    {args[1]},
			"quantity": {args[2]}})
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
