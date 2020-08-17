package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

type assetsMessage struct {
	Assets map[string]map[string]float64
}

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Retrieve assets",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/assets")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var am assetsMessage

		json.Unmarshal(body, &am)
		if err != nil {
			panic(err)
		}

		for venue, balances := range am.Assets {
			fmt.Println(venue)
			for asset, quantity := range balances {
				fmt.Printf("\t%s: %f\n", asset, quantity)
			}
		}
	},
}

var setAssetsCmd = &cobra.Command{
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
