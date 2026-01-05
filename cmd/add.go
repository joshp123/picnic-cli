package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <product_id> [count]",
		Short: "Add a product to the cart",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			count := 1
			if len(args) > 1 {
				parsed, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid count: %s", args[1])
				}
				count = parsed
			}
			client, err := getClient()
			if err != nil {
				return err
			}
			cart, err := client.AddToCart(args[0], count)
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Printf("\u2705 Added %dx product %s to cart\n", count, args[0])
			showCartSummary(cart)
			return nil
		},
	}
	return cmd
}
