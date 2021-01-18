package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/spf13/cobra"
)

type pricesMessage struct {
	Prices map[string]float64
}

func (pm *pricesMessage) assets() []string {
	assets := make([]string, len(pm.Prices))
	i := 0
	for asset, _ := range pm.Prices {
		assets[i] = asset
		i++
	}
	sort.Strings(assets)

	return assets
}

var pricesCmd = &cobra.Command{
	Use:   "prices",
	Short: "Retrieve current prices",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/prices")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var pm pricesMessage

		json.Unmarshal(body, &pm)
		if err != nil {
			panic(err)
		}

		for _, asset := range pm.assets() {
			fmt.Printf("%s: %f\n", asset, pm.Prices[asset])
		}
	},
}
