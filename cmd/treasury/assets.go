package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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

func (am *assetsMessage) hasAssets(venue string) bool {
	if len(am.Assets[venue]) == 0 {
		return false
	}

	total := 0.0
	for _, quantity := range am.Assets[venue] {
		total += quantity
	}

	return total > 0
}

func (am *assetsMessage) venueAssets(venue string) []assetQuantity {
	assets := []string{}
	for asset, quantity := range am.Assets[venue] {
		if quantity != 0 {
			assets = append(assets, asset)
		}
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
			if !am.hasAssets(venue) {
				continue
			}
			fmt.Println(venue)
			for _, aq := range am.venueAssets(venue) {
				fmt.Printf("\t%s: %.8f\n", aq.asset, aq.quantity)
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
