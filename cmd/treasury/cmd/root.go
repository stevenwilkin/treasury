package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "treasury",
		Short: "CLI interface to treasuryd",
	}
)

func Execute() error {
	return rootCmd.Execute()
}
