package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "treasury",
	Short: "CLI interface to treasuryd",
}

func init() {
	rootCmd.AddCommand(pricesCmd)
	rootCmd.AddCommand(assetsCmd)
	rootCmd.AddCommand(costCmd)
	rootCmd.AddCommand(pnlCmd)
	rootCmd.AddCommand(alertsCmd)
	rootCmd.AddCommand(fundingCmd)

	assetsCmd.AddCommand(setAssetsCmd)
	alertsCmd.AddCommand(alertsPriceCmd, alertsClearCmd, alertsFundingCmd)
}
