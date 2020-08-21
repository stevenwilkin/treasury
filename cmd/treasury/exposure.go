package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type exposureMessage struct {
	Value float64 `json:"value"`
}

var exposureCmd = &cobra.Command{
	Use:   "exposure",
	Short: "Retrieve BTC long exposure",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/exposure")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var em exposureMessage

		json.Unmarshal(body, &em)
		if err != nil {
			panic(err)
		}

		fmt.Println(em.Value)
	},
}
