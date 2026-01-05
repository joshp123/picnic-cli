package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for products",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			results, err := searchArticlesRaw(query)
			if err != nil {
				invalidateAuthCache()
				return err
			}
			if len(results) == 0 {
				fmt.Printf("No products found for %q\n", query)
				return nil
			}

			fmt.Printf("\U0001F50D Search results for %q:\n\n", query)
			limit := 10
			if len(results) < limit {
				limit = len(results)
			}
			for i := 0; i < limit; i++ {
				item := results[i]
				price := formatPrice(item.PriceIncludingPromotions())
				unit := strings.TrimSpace(item.UnitQuantity)
				fmt.Printf("%d. [%s] %s\n", i+1, item.Id, item.Name)
				if unit != "" {
					fmt.Printf("   %s | %s\n", price, unit)
				} else {
					fmt.Printf("   %s\n", price)
				}
				if item.ImageId != "" {
					fmt.Printf("   https://storefront-prod.nl.picnicinternational.com/static/images/%s/medium.png\n", item.ImageId)
				}
				fmt.Println()
			}
			return nil
		},
	}
	return cmd
}
