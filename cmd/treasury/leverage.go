package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type leverageMessage struct {
	Value float64 `json:"value"`
}

var leverageCmd = &cobra.Command{
	Use:   "leverage",
	Short: "Retrieve account leverage",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/leverage")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var em leverageMessage

		json.Unmarshal(body, &em)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%.2f\n", em.Value)
	},
}
