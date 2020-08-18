package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type fundingMessage struct {
	Current   float64 `json:"current"`
	Predicted float64 `json:"predicted"`
}

var fundingCmd = &cobra.Command{
	Use:   "funding",
	Short: "Retrieve funding",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/funding")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var pm fundingMessage

		json.Unmarshal(body, &pm)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Current:   %f%%\n", pm.Current*100)
		fmt.Printf("Predicted: %f%%\n", pm.Predicted*100)
	},
}
