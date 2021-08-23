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
	rootCmd.AddCommand(exposureCmd)
	rootCmd.AddCommand(leverageCmd)
	rootCmd.AddCommand(sizeCmd)
	rootCmd.AddCommand(feedsCmd)
	rootCmd.AddCommand(indicatorsCmd)
	rootCmd.AddCommand(loanCmd)

	assetsCmd.AddCommand(setAssetsCmd)
	alertsCmd.AddCommand(
		alertsPriceCmd, alertsClearCmd, alertsFundingCmd, alertsLeverageCmd)
	pnlCmd.AddCommand(pnlUsdCmd)
	sizeCmd.AddCommand(sizeUpdateCmd)
	feedsCmd.AddCommand(feedsReactivateCmd)
	loanCmd.AddCommand(loanSetCmd)
}
