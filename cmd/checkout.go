package cmd

import (
	"fmt"
	"strings"

	picnic "github.com/simonmartyr/picnic-api"
	"github.com/spf13/cobra"
)

func checkoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout",
		Short: "Checkout and payment",
	}
	cmd.AddCommand(checkoutStartCmd())
	cmd.AddCommand(checkoutStatusCmd())
	cmd.AddCommand(checkoutCancelCmd())
	cmd.AddCommand(checkoutPayCmd())
	return cmd
}

func checkoutStartCmd() *cobra.Command {
	var resolveKey string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start checkout for the current cart",
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
			if cart.TotalCount == 0 {
				return fmt.Errorf("cart is empty")
			}

			var checkout *picnic.Checkout
			var cerr *picnic.CheckoutError
			if strings.TrimSpace(resolveKey) != "" {
				checkout, cerr = client.CheckoutWithResolveKey(cart.Mts, resolveKey)
			} else {
				checkout, cerr = client.StartCheckout(cart.Mts)
			}
			if cerr != nil {
				fmt.Printf("Checkout error: %s\n", cerr.Error())
				if cerr.Title != "" || cerr.Message != "" {
					fmt.Printf("%s - %s\n", cerr.Title, cerr.Message)
				}
				if cerr.ResolveKey != "" {
					fmt.Printf("Resolve key required: %s\n", cerr.ResolveKey)
				}
				if cerr.Blocking {
					fmt.Println("Blocking: true")
				}
				return nil
			}
			fmt.Printf("Checkout started. Order ID: %s\n", checkout.OrderId)
			fmt.Printf("Total: %s | Items: %d\n", formatPrice(checkout.TotalPrice), checkout.TotalCount)
			return nil
		},
	}
	cmd.Flags().StringVar(&resolveKey, "resolve", "", "Resolve key if required (e.g., age_verified)")
	return cmd
}

func checkoutStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <transaction_id>",
		Short: "Check checkout status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			status, err := client.GetCheckoutStatus(args[0])
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Printf("Checkout status: %s\n", status)
			return nil
		},
	}
	return cmd
}

func checkoutCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <transaction_id>",
		Short: "Cancel checkout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if err := client.CancelCheckout(args[0]); err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Println("Checkout cancelled")
			return nil
		},
	}
	return cmd
}

func checkoutPayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pay <order_id>",
		Short: "Initiate payment for an order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			payment, err := client.InitiatePayment(args[0])
			if err != nil {
				invalidateAuthCache()
				return err
			}
			fmt.Printf("Payment initiated. Transaction ID: %s\n", payment.TransactionId)
			if payment.IssuerAuthenticationUrl != "" {
				fmt.Printf("Issuer authentication URL: %s\n", payment.IssuerAuthenticationUrl)
			}
			if payment.Action.RedirectUrl != "" {
				fmt.Printf("Redirect URL: %s\n", payment.Action.RedirectUrl)
			}
			return nil
		},
	}
	return cmd
}
