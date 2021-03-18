package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type indicatorsMessage struct {
	THBPremium  float64 `json:"thb_premium"`
	USDTPremium float64 `json:"usdt_premium"`
}

var indicatorsCmd = &cobra.Command{
	Use:   "indicators",
	Short: "Retrieve indicators",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/indicators")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var im indicatorsMessage

		json.Unmarshal(body, &im)
		if err != nil {
			panic(err)
		}

		fmt.Printf("THB  Premium: %.2f%%\n", im.THBPremium*100)
		fmt.Printf("USDT Premium: %.2f%%\n", im.USDTPremium*100)
	},
}
