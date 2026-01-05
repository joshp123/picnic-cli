package cmd

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	picnic "github.com/simonmartyr/picnic-api"
)

type authCache struct {
	AuthKey   string `json:"authKey"`
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
}

type loginInput struct {
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	ClientId int    `json:"client_id"`
}

func getClient() (*picnic.Client, error) {
	ctx, err := getAuthContext()
	if err != nil {
		return nil, err
	}

	client := picnic.New(http.DefaultClient, picnic.WithCountry(ctx.Country), picnic.WithToken(ctx.Token))
	return client, nil
}

type authContext struct {
	Token   string
	Country string
	Email   string
}

func getAuthContext() (authContext, error) {
	email := strings.TrimSpace(os.Getenv("PICNIC_EMAIL"))
	password := os.Getenv("PICNIC_PASSWORD")
	country := strings.TrimSpace(os.Getenv("PICNIC_COUNTRY"))
	if country == "" {
		country = "NL"
	}

	tokenPath, err := tokenFilePath()
	if err != nil {
		return authContext{}, err
	}

	var token string
	if tokenPath != "" {
		cached, err := loadAuthCache(tokenPath)
		if err == nil && cached.AuthKey != "" && cached.Email == email {
			token = cached.AuthKey
		}
	}

	if token == "" {
		if email == "" || password == "" {
			fileEmail, filePassword, err := readCredentialsFile()
			if err != nil {
				return authContext{}, err
			}
			if email == "" {
				email = fileEmail
			}
			if password == "" {
				password = filePassword
			}
		}
		if email == "" || password == "" {
			return authContext{}, fmt.Errorf("PICNIC_EMAIL and PICNIC_PASSWORD must be set, or provide PICNIC_AUTH_FILE")
		}
		loginToken, err := login(country, email, password)
		if err != nil {
			return authContext{}, err
		}
		token = loginToken
		if tokenPath != "" {
			_ = saveAuthCache(tokenPath, authCache{
				AuthKey:   token,
				Email:     email,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}

	return authContext{
		Token:   token,
		Country: country,
		Email:   email,
	}, nil
}

func login(country, email, password string) (string, error) {
	baseURL := fmt.Sprintf("https://storefront-prod.%s.picnicinternational.com/api/15", strings.ToLower(country))
	url := baseURL + "/user/login"
	body, err := json.Marshal(loginInput{
		Key:      email,
		Secret:   md5Hash(password),
		ClientId: 20100,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("login failed: %s", strings.TrimSpace(string(payload)))
	}

	token := res.Header.Get("x-picnic-auth")
	if token == "" {
		return "", fmt.Errorf("login failed: missing auth token")
	}
	return token, nil
}

func tokenFilePath() (string, error) {
	if v := strings.TrimSpace(os.Getenv("PICNIC_TOKEN_FILE")); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".picnic-token"), nil
}

func historyFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".picnic-history.json"), nil
}

func preferencesFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".picnic-preferences.json"), nil
}

func loadAuthCache(path string) (authCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return authCache{}, err
	}
	var cached authCache
	if err := json.Unmarshal(data, &cached); err != nil {
		return authCache{}, err
	}
	return cached, nil
}

func saveAuthCache(path string, cache authCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func invalidateAuthCache() {
	path, err := tokenFilePath()
	if err != nil {
		return
	}
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path)
	}
}

func md5Hash(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

func formatPrice(cents int) string {
	if cents == 0 {
		return "?"
	}
	return fmt.Sprintf("\u20ac%.2f", float64(cents)/100)
}

func readCredentialsFile() (string, string, error) {
	path := strings.TrimSpace(os.Getenv("PICNIC_AUTH_FILE"))
	if path == "" {
		return "", "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read PICNIC_AUTH_FILE: %w", err)
	}
	email, password := parseCredentials(string(data))
	if email == "" || password == "" {
		return "", "", fmt.Errorf("PICNIC_AUTH_FILE is missing email or password")
	}
	return email, password, nil
}

func parseCredentials(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var email string
	var password string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "email=") || strings.HasPrefix(lower, "username=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				email = strings.TrimSpace(parts[1])
			}
			continue
		}
		if strings.HasPrefix(lower, "password=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				password = strings.TrimSpace(parts[1])
			}
			continue
		}
	}

	if email != "" && password != "" {
		return email, password
	}

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.EqualFold(line, "[username]") || strings.EqualFold(line, "[email]") {
			for j := i + 1; j < len(lines); j++ {
				candidate := strings.TrimSpace(lines[j])
				if candidate != "" {
					email = candidate
					break
				}
			}
		}
		if strings.EqualFold(line, "[password]") {
			for j := i + 1; j < len(lines); j++ {
				candidate := strings.TrimSpace(lines[j])
				if candidate != "" {
					password = candidate
					break
				}
			}
		}
	}

	if email != "" && password != "" {
		return email, password
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if email == "" {
			email = trimmed
		} else if password == "" {
			password = trimmed
			break
		}
	}

	return email, password
}
