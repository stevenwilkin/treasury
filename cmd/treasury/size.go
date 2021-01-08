package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

type sizeMessage struct {
	Size int `json:"size"`
}

var sizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Retrieve size",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/size")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var pm sizeMessage

		json.Unmarshal(body, &pm)
		if err != nil {
			panic(err)
		}

		fmt.Println(pm.Size)
	},
}

var sizeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update size",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/size/update")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var pm sizeMessage

		json.Unmarshal(body, &pm)
		if err != nil {
			panic(err)
		}

		fmt.Println(pm.Size)
	},
}
