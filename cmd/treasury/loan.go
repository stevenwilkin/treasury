package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

type loanMessage struct {
	Loan float64 `json:"loan"`
}

var loanCmd = &cobra.Command{
	Use:   "loan",
	Short: "Retrieve outstanding loan",
	Run: func(cmd *cobra.Command, args []string) {
		var lm loanMessage
		get("/loan", &lm)

		fmt.Println(lm.Loan)
	},
}

var loanSetCmd = &cobra.Command{
	Use:   "set [loan]",
	Short: "Set outstanding loan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.PostForm("http://unix/loan/set", url.Values{
			"loan": {args[0]}})
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
