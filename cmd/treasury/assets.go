package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
