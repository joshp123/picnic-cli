package cmd

import (
	"fmt"

	picnic "github.com/simonmartyr/picnic-api"
	"github.com/spf13/cobra"
)

func cartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cart",
		Short: "View the shopping cart",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			cart, err := client.GetCart()
			if err != nil {
				invalidateAuthCache()
				return err
			}
			showCart(cart)
			return nil
		},
	}
	return cmd
}

func showCartSummary(cart *picnic.Order) {
	if cart == nil {
		return
	}
	items := cart.TotalCount
	total := formatPrice(cart.TotalPrice)
	fmt.Printf("\n\U0001F6D2 Cart: %d items | Total: %s\n", items, total)
}

func showCart(cart *picnic.Order) {
	if cart == nil || len(cart.Items) == 0 {
		fmt.Println("\U0001F6D2 Cart is empty")
		return
	}

	fmt.Println("\U0001F6D2 Shopping Cart:\n")
	for _, line := range cart.Items {
		if len(line.Items) == 0 {
			continue
		}
		for _, article := range line.Items {
			if article.Name == "" {
				continue
			}
			qty := article.Quantity()
			if qty == 0 {
				qty = 1
			}
			price := formatPrice(article.DisplayPrice)
			fmt.Printf("  %dx %s %s\n", qty, article.Name, price)
		}
	}

	showCartSummary(cart)
}
