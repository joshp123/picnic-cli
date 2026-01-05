package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func debugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug deliveries API output",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if store, err := client.GetMyStore(); err != nil {
				fmt.Printf("MyStore error: %s\n", err)
			} else if store != nil {
				fmt.Printf("MyStore id: %s\n", store.Id)
				fmt.Printf("MyStore first_time_user: %v\n", store.FirstTimeUser)
				fmt.Printf("MyStore catalog count: %d\n", len(store.Catalog))
				fmt.Printf("MyStore landing_page_hit: %s\n", store.LandingPageHit)
			}
			deliveries, err := client.GetDeliveries(nil)
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Printf("Total deliveries: %d\n", len(*deliveries))
			if len(*deliveries) == 0 {
				return nil
			}
			fmt.Println("\nFirst delivery:")
			first, _ := json.MarshalIndent((*deliveries)[0], "", "  ")
			fmt.Println(string(first))

			detailID := (*deliveries)[0].DeliveryId
			if detailID == "" {
				detailID = (*deliveries)[0].Id
			}
			fmt.Printf("\nTrying detail ID: %s\n", detailID)
			if detailID == "" {
				return nil
			}

			detail, err := client.GetDelivery(detailID)
			if err != nil {
				fmt.Printf("Detail error: %s\n", err)
				return nil
			}
			fmt.Printf("\nDetail orders: %d\n", len(detail.Orders))
			if len(detail.Orders) > 0 && len(detail.Orders[0].Items) > 0 {
				fmt.Println("First order items sample:")
				sample := detail.Orders[0].Items
				if len(sample) > 2 {
					sample = sample[:2]
				}
				payload, _ := json.MarshalIndent(sample, "", "  ")
				fmt.Println(string(payload))
			}
			return nil
		},
	}
	return cmd
}
