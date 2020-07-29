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
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(costCmd)
	rootCmd.AddCommand(pnlCmd)
}
