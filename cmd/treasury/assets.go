package main

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/spf13/cobra"
)

type assetsMessage struct {
	Assets map[string]map[string]float64
}

type assetQuantity struct {
	asset    string
	quantity float64
}

func (am *assetsMessage) venues() []string {
	result := make([]string, len(am.Assets))
	i := 0
	for venue, _ := range am.Assets {
		result[i] = venue
		i++
	}
	sort.Strings(result)

	return result
}

func (am *assetsMessage) venueAssets(venue string) []assetQuantity {
	assets := []string{}
	for asset, _ := range am.Assets[venue] {
		assets = append(assets, asset)
	}
	sort.Strings(assets)

	results := []assetQuantity{}
	for _, asset := range assets {
		results = append(results, assetQuantity{
			asset:    asset,
			quantity: am.Assets[venue][asset]})
	}
	return results
}

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Retrieve assets",
	Run: func(cmd *cobra.Command, args []string) {
		var am assetsMessage
		get("/assets", &am)

		for _, venue := range am.venues() {
			fmt.Println(venue)
			for _, aq := range am.venueAssets(venue) {
				if aq.asset == "BTC" {
					fmt.Printf("\t%s: %.8f\n", aq.asset, aq.quantity)
				} else {
					fmt.Printf("\t%s: %.2f\n", aq.asset, aq.quantity)
				}
			}
		}
	},
}

var setAssetsCmd = &cobra.Command{
	Use:   "set [venue] [asset] [quantity]",
	Short: "Set quantity of an asset within a venue",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		post("/set", url.Values{
			"venue":    {args[0]},
			"asset":    {args[1]},
			"quantity": {args[2]}})
	},
}
