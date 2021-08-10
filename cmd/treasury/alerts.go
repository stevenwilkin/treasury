package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

type alertsMessage struct {
	Active      bool
	Description string
}

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Retrieve alerts",
	Run: func(cmd *cobra.Command, args []string) {
		var am []alertsMessage
		get("/alerts", &am)

		for _, alert := range am {
			active := "Active  "
			if !alert.Active {
				active = "Inactive"
			}
			fmt.Printf("%s - %s\n", active, alert.Description)
		}
	},
}

var alertsPriceCmd = &cobra.Command{
	Use:   "price [value]",
	Short: "Set price alert",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.PostForm("http://unix/alerts/price", url.Values{
			"value": {args[0]}})
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

var alertsFundingCmd = &cobra.Command{
	Use:   "funding",
	Short: "Set funding alert",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/alerts/funding")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	},
}

var alertsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear alerts",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get("http://unix/alerts/clear")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	},
}
