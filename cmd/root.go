package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "picnic",
	Short:        "Picnic CLI for managing your grocery cart",
	SilenceUsage: true,
}

func Execute() {
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(cartCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(debugCmd())
	rootCmd.AddCommand(slotsCmd())
	rootCmd.AddCommand(slotCmd())
	rootCmd.AddCommand(checkoutCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
