package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/spf13/cobra"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

type pricesMessage struct {
	Prices map[string]float64
}

var pricesCmd = &cobra.Command{
	Use:   "prices",
	Short: "Retrieve current prices",
	Run: func(cmd *cobra.Command, args []string) {
		client := http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		}

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

func init() {
	rootCmd.AddCommand(pricesCmd)
}
