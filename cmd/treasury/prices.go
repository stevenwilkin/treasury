package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type pricesMessage struct {
	Prices map[string]float64
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

		for asset, price := range pm.Prices {
			fmt.Printf("%s: %f\n", asset, price)
		}
	},
}
