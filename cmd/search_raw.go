package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	picnic "github.com/simonmartyr/picnic-api"
)

type tokenMeta struct {
	PcClid int    `json:"pc:clid"`
	PcDid  string `json:"pc:did"`
}

const appVersion = "1.15.243-18832"

func searchArticlesRaw(query string) ([]picnic.SingleArticle, error) {
	ctx, err := getAuthContext()
	if err != nil {
		return nil, err
	}
	meta, err := parseTokenMeta(ctx.Token)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("https://storefront-prod.%s.picnicinternational.com/api/15/pages/search-page-results?search_term=%s",
		strings.ToLower(ctx.Country), url.QueryEscape(query))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-picnic-auth", ctx.Token)
	req.Header.Set("x-picnic-agent", fmt.Sprintf("%d;%s;", meta.PcClid, appVersion))
	req.Header.Set("x-picnic-did", meta.PcDid)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("search failed: status %d", resp.StatusCode)
	}

	var payload interface{}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}

	var results []picnic.SingleArticle
	extractSellingUnits(payload, &results)
	return results, nil
}

func parseTokenMeta(token string) (tokenMeta, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenMeta{}, fmt.Errorf("invalid token structure")
	}
	payload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenMeta{}, err
	}
	var meta tokenMeta
	if err := json.Unmarshal(payload, &meta); err != nil {
		return tokenMeta{}, err
	}
	if meta.PcClid == 0 || meta.PcDid == "" {
		return tokenMeta{}, fmt.Errorf("token missing pc:clid or pc:did")
	}
	return meta, nil
}

func extractSellingUnits(node interface{}, out *[]picnic.SingleArticle) {
	switch v := node.(type) {
	case map[string]interface{}:
		if content, ok := v["content"].(map[string]interface{}); ok {
			if typ, ok := content["type"].(string); ok && typ == "SELLING_UNIT_TILE" {
				if sellingUnit, ok := content["sellingUnit"]; ok {
					buf, err := json.Marshal(sellingUnit)
					if err == nil {
						var article picnic.SingleArticle
						if err := json.Unmarshal(buf, &article); err == nil {
							if strings.TrimSpace(article.Id) != "" {
								*out = append(*out, article)
							}
						}
					}
				}
			}
		}
		for _, val := range v {
			extractSellingUnits(val, out)
		}
	case []interface{}:
		for _, val := range v {
			extractSellingUnits(val, out)
		}
	}
}
