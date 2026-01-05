package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func slotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slots",
		Short: "List available delivery slots",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			slots, err := client.GetDeliverySlots()
			if err != nil {
				invalidateAuthCache()
				return err
			}
			if slots == nil || len(slots.DeliverySlots) == 0 {
				fmt.Println("No delivery slots found")
				return nil
			}

			fmt.Println("Delivery slots:\n")
			for _, slot := range slots.DeliverySlots {
				status := "unavailable"
				if slot.IsAvailable {
					status = "available"
				}
				selected := ""
				if slot.Selected {
					selected = " (selected)"
				}
				window := strings.TrimSpace(slot.WindowStart + " - " + slot.WindowEnd)
				fmt.Printf("- %s%s\n  id: %s | %s\n", window, selected, slot.SlotId, status)
				if slot.MinimumOrderValue > 0 {
					fmt.Printf("  minimum order: %s\n", formatPrice(slot.MinimumOrderValue))
				}
				if !slot.IsAvailable && strings.TrimSpace(slot.UnavailabilityReason) != "" {
					fmt.Printf("  reason: %s\n", slot.UnavailabilityReason)
				}
				fmt.Println()
			}
			return nil
		},
	}
	return cmd
}

func slotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slot",
		Short: "Manage delivery slots",
	}
	cmd.AddCommand(slotSetCmd())
	return cmd
}

func slotSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <slot_id>",
		Short: "Select a delivery slot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			order, err := client.SetDeliverySlot(args[0])
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Printf("Selected slot %s\n", args[0])
			showCartSummary(order)
			return nil
		},
	}
	return cmd
}
