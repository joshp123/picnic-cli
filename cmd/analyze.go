package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	picnic "github.com/simonmartyr/picnic-api"
	"github.com/spf13/cobra"
)

type productEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Unit     string `json:"unit"`
	Quantity int    `json:"quantity"`
	Date     string `json:"date"`
}

type productCount struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Unit          string `json:"unit"`
	Price         int    `json:"price"`
	Count         int    `json:"count"`
	TotalQuantity int    `json:"totalQuantity"`
}

type categoryPreference struct {
	Default      productCount   `json:"default"`
	Alternatives []productCount `json:"alternatives"`
}

func analyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze-orders",
		Short: "Analyze order history and infer preferences",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			products, err := fetchAllOrders(client)
			if err != nil {
				invalidateAuthCache()
				return err
			}
			if len(products) == 0 {
				fmt.Println("No products found in order history")
				return nil
			}

			categories, preferences, topProducts := analyzePreferences(products)
			showAnalysis(categories, preferences, topProducts)
			return nil
		},
	}
	return cmd
}

func fetchAllOrders(client interface {
	GetDeliveries(filter []picnic.DeliveryStatus) (*[]picnic.Delivery, error)
	GetDelivery(deliveryId string) (*picnic.Delivery, error)
}) ([]productEntry, error) {
	fmt.Println("\U0001F4E6 Fetching order history...\n")

	deliveries, err := client.GetDeliveries(nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Found %d deliveries\n\n", len(*deliveries))

	var allProducts []productEntry
	processed := 0

	recent := *deliveries
	if len(recent) > 50 {
		recent = recent[:50]
	}

	for _, delivery := range recent {
		processed++
		fmt.Printf("\rProcessing %d/%d...", processed, len(recent))

		deliveryID := delivery.DeliveryId
		if deliveryID == "" {
			deliveryID = delivery.Id
		}
		if deliveryID == "" {
			continue
		}

		date := delivery.CreationTime
		if delivery.DeliveryTime.Start != "" {
			date = delivery.DeliveryTime.Start
		}

		detail, err := client.GetDelivery(deliveryID)
		if err != nil {
			continue
		}

		for _, order := range detail.Orders {
			for _, line := range order.Items {
				for _, article := range line.Items {
					if article.Type != "ORDER_ARTICLE" || article.Id == "" || article.Name == "" {
						continue
					}
					qty := article.Quantity()
					if qty == 0 {
						qty = 1
					}
					price := line.DisplayPrice
					if price == 0 {
						price = line.Price
					}
					allProducts = append(allProducts, productEntry{
						ID:       article.Id,
						Name:     article.Name,
						Price:    price,
						Unit:     article.UnitQuantity,
						Quantity: qty,
						Date:     date,
					})
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n\n\u2705 Extracted %d product entries\n", len(allProducts))

	historyPath, err := historyFilePath()
	if err == nil {
		if err := writeJSONFile(historyPath, allProducts); err == nil {
			fmt.Printf("Saved to %s\n", historyPath)
		}
	}

	return allProducts, nil
}

func analyzePreferences(products []productEntry) (map[string][]productCount, map[string]categoryPreference, []productCount) {
	counts := map[string]*productCount{}
	for _, p := range products {
		entry, ok := counts[p.ID]
		if !ok {
			counts[p.ID] = &productCount{
				ID:    p.ID,
				Name:  p.Name,
				Unit:  p.Unit,
				Price: p.Price,
			}
			entry = counts[p.ID]
		}
		entry.Count++
		entry.TotalQuantity += p.Quantity
	}

	sorted := make([]productCount, 0, len(counts))
	for _, v := range counts {
		sorted = append(sorted, *v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	categories := map[string][]productCount{
		"melk":        {},
		"boter":       {},
		"brood":       {},
		"kaas":        {},
		"eieren":      {},
		"yoghurt":     {},
		"vleeswaren":  {},
		"fruit":       {},
		"groente":     {},
		"vlees":       {},
		"drank":       {},
		"snoep":       {},
		"diepvries":   {},
		"overig":      {},
	}

	patterns := map[string]*regexp.Regexp{
		"melk":       regexp.MustCompile("(?i)melk|milk|havermelk|amandel.*melk|soja.*melk|sojamelk|oatly|alpro|verse\\s?melk|volle\\s?melk|halfvolle\\s?melk"),
		"boter":      regexp.MustCompile("(?i)boter|margarine|halvarine|becel|rama"),
		"brood":      regexp.MustCompile("(?i)brood|bol(len)?|broodje|toast|croissant|baguette|ciabatta|pistolet|boterham"),
		"kaas":       regexp.MustCompile("(?i)kaas|cheese|gouda|emmentaler|mozzarella|parmezaan|parmesan|feta|camembert|brie|plak"),
		"eieren":     regexp.MustCompile("(?i)\\bei(er|ren)\\b|\\begg(s)?\\b|vrije\\s?uitloop|scharrel"),
		"yoghurt":    regexp.MustCompile("(?i)yoghurt|yogurt|kwark|skyr|pudding|dessert"),
		"vleeswaren": regexp.MustCompile("(?i)ham|salami|worst|vleeswaren|bacon|spek|mortadella|leverworst"),
		"fruit":      regexp.MustCompile("(?i)appel|banaan|sinaasappel|peer|druif|bessen|mango|ananas|kiwi|citroen|limoen|avocado|meloen"),
		"groente":    regexp.MustCompile("(?i)tomaat|komkommer|paprika|ui|wortel|sla|spinazie|broccoli|courgette|aardappel|champignon|prei"),
		"vlees":      regexp.MustCompile("(?i)kip|kalf|rund|runder|varken|gehakt|filet|steak|schnitzel|goulash|shoarma"),
		"drank":      regexp.MustCompile("(?i)water|sap|cola|limonade|fanta|sprite|bier|wijn|thee|koffie|energy|spa|frisdrank|sapjes"),
		"snoep":      regexp.MustCompile("(?i)chocolade|koek|cookie|gummi|chips|snack|reep|ijs|bonbon|snoep"),
		"diepvries":  regexp.MustCompile("(?i)diepvries|vries|pizza|patat|fri(et|t)en|vissticks|spinazie.*vries"),
	}

	for _, product := range sorted {
		categorized := false
		for cat, pattern := range patterns {
			if pattern.MatchString(product.Name) {
				categories[cat] = append(categories[cat], product)
				categorized = true
				break
			}
		}
		if !categorized {
			categories["overig"] = append(categories["overig"], product)
		}
	}

	preferences := map[string]categoryPreference{}
	for cat, items := range categories {
		if cat == "overig" || len(items) == 0 {
			continue
		}
		pref := categoryPreference{
			Default:      items[0],
			Alternatives: []productCount{},
		}
		if len(items) > 1 {
			max := 5
			if len(items) < max+1 {
				max = len(items) - 1
			}
			pref.Alternatives = append(pref.Alternatives, items[1:max+1]...)
		}
		preferences[cat] = pref
	}

	prefPath, err := preferencesFilePath()
	if err == nil {
		_ = writeJSONFile(prefPath, preferences)
		fmt.Printf("\n\u2705 Saved preferences to %s\n", prefPath)
	}

	topProducts := sorted
	if len(sorted) > 30 {
		topProducts = sorted[:30]
	}

	return categories, preferences, topProducts
}

func showAnalysis(categories map[string][]productCount, preferences map[string]categoryPreference, topProducts []productCount) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\U0001F4CA JOUW WINKELGEWOONTES")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n\U0001F3C6 TOP 15 MEEST GEKOCHTE PRODUCTEN:\n")
	limit := 15
	if len(topProducts) < limit {
		limit = len(topProducts)
	}
	for i := 0; i < limit; i++ {
		p := topProducts[i]
		price := formatPrice(p.Price)
		fmt.Printf("%2d. %s (%dx) %s\n", i+1, p.Name, p.Count, price)
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("\U0001F3F7\ufe0f  STANDAARDPRODUCTEN PER CATEGORIE:\n")

	emojis := map[string]string{
		"melk": "\U0001F95B", "boter": "\U0001F9C8", "brood": "\U0001F35E", "kaas": "\U0001F9C0", "eieren": "\U0001F95A",
		"yoghurt": "\U0001F944", "vleeswaren": "\U0001F953", "fruit": "\U0001F34E", "groente": "\U0001F955",
		"vlees": "\U0001F356", "drank": "\U0001F964", "snoep": "\U0001F36B", "diepvries": "\U0001F9CA",
	}

	for cat, prefs := range preferences {
		emoji := emojis[cat]
		if emoji == "" {
			emoji = "\U0001F4E6"
		}
		fmt.Printf("%s %s\n", emoji, strings.ToUpper(cat))
		fmt.Printf("   -> %s\n", prefs.Default.Name)
		fmt.Printf("     ID: %s | %dx gekauft\n", prefs.Default.ID, prefs.Default.Count)
		if len(prefs.Alternatives) > 0 {
			alts := make([]string, 0, len(prefs.Alternatives))
			for i, alt := range prefs.Alternatives {
				if i >= 2 {
					break
				}
				alts = append(alts, alt.Name)
			}
			fmt.Printf("     Alternativen: %s\n", strings.Join(alts, ", "))
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\u2705 Nu weet ik wat je bedoelt met \"koop melk\" enz.!")
	fmt.Println(strings.Repeat("=", 60))

}

func writeJSONFile(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
