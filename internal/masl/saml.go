package masl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/Sirupsen/logrus"
	"fmt"
	"os"
)

const (
	generateTokenAPI = "auth/oauth2/token"
	samlAssertionAPI = "api/1/saml_assertion"
)

// APITokenResponse represents the JSON response from the OneLogin Generate APITokenResponse REST call
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

type SAMLAssertionRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	AppID           string `json:"app_id"`
	Subdomain       string `json:"subdomain"`
}

type SAMLAssertionResponse struct {
	Status struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   bool   `json:"error"`
	} `json:"status"`
	Data []struct {
		CallbackURL string `json:"callback_url"`
		Devices     []struct {
			DeviceID   int    `json:"device_id"`
			DeviceType string `json:"device_type"`
		} `json:"devices"`
		StateToken string `json:"state_token"`
		User       struct {
			Email     string `json:"email"`
			Lastname  string `json:"lastname"`
			Username  string `json:"username"`
			ID        int    `json:"id"`
			Firstname string `json:"firstname"`
		} `json:"user"`
	} `json:"data"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func logRequest(log *logrus.Logger, req *http.Request) {
	dump, _ := httputil.DumpRequest(req, true)
	log.Debug(string(dump))
}

func logResponse(log *logrus.Logger, resp *http.Response) {
	dump, _ := httputil.DumpResponse(resp, true)
	log.Debug(string(dump))
}

// GenerateToken: Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
// Generate an access token and refresh token to access onelogin's resource APIs.
func GenerateToken(conf Config, log *logrus.Logger) string {

	url := conf.BaseURL + generateTokenAPI

	var jsonStr = []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	auth := "client_id:" + conf.ClientID + ",client_secret:" + conf.ClientSecret
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	logRequest(log, req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	logResponse(log, resp)

	apiToken := APITokenResponse{}
	json.NewDecoder(resp.Body).Decode(&apiToken)

	//TODO: being a bit optimistic here ;)
	return apiToken.Data[0].AccessToken
}

// SAMLAssertion: Call to https://api.eu.onelogin.com/api/1/saml_assertion
func SAMLAssertion(conf Config, log *logrus.Logger, password string, apiToken string) {

	url := conf.BaseURL + samlAssertionAPI

	requestBody, err := json.Marshal(SAMLAssertionRequest{
		UsernameOrEmail: conf.Username,
		Password:        password,
		AppID:           conf.AppID,
		Subdomain:       conf.Subdomain})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}
	auth := "bearer:" + apiToken
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	logRequest(log, req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	logResponse(log, resp)

	samlAssertion := SAMLAssertionResponse{}
	json.NewDecoder(resp.Body).Decode(&samlAssertion)

	if samlAssertion.Status.Code == 401{
		fmt.Print("Password doesn't match :(")
		os.Exit(0)
	}
}
