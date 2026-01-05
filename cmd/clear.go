package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func clearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the shopping cart",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			_, err = client.ClearCart()
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Println("\U0001F5D1 Cart cleared")
			return nil
		},
	}
	return cmd
}
