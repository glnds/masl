package masl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Client struct {
	c            *http.Client
	baseURL      string
	clientID     string
	clientSecret string
}

// APITokenResponse represents the OneLogin Generate API Token response
type APITokenResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		AccessToken  string    `json:"access_token"`
		CreatedAt    time.Time `json:"created_at"`
		ExpiresIn    int       `json:"expires_in"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		AccountID    int       `json:"account_id"`
	} `json:"data"`
}

func New(config Config) *Client {
	client := &http.Client{Timeout: 10 * time.Second}

	return &Client{
		c:            client,
		baseURL:      config.BaseURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
	}
}

// GenerateToken Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
func (client *Client) GenerateToken() (string, error) {

	url := client.baseURL + "auth/oauth2/token"
	requestBody := []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	auth := "client_id:" + client.clientID + ",client_secret:" + client.clientSecret
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.c.Do(req)
	if err != nil {
		return "", err
	}

	// deserialize the response and return our weather data
	defer res.Body.Close()
	var apiToken APITokenResponse
	if err := json.NewDecoder(res.Body).Decode(&apiToken); err != nil {
		return "", err
	}
	return apiToken.Data[0].AccessToken, nil
}
