package masl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	generateTokenAPI = "auth/oauth2/token"
)

// Token is a onelogin generated token
type Token struct {
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

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(req *http.Request, target interface{}) error {

	resp, err := httpClient.Do(req)
	if err != nil {
		return (err)
	}
	defer resp.Body.Close()
	debugREST(httputil.DumpResponse(resp, true))

	return json.NewDecoder(resp.Body).Decode(target)
}

// Login authenticate against OneLogin.
func Login(conf Config, log *logrus.Logger) {
	generateToken(conf)
}

// Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
// Generate an access token and refresh token to access onelogin's resource APIs.
func generateToken(conf Config) {

	auth := ("client_id:" + conf.ClientID + ",client_secret:" + conf.ClientSecret)
	url := conf.BaseURL + generateTokenAPI

	var jsonStr = []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	debugREST(httputil.DumpRequestOut(req, true))

	test := Token{}
	getJSON(req, &test)

	fmt.Println(test.Data)
}

func debugREST(data []byte, err error) {
	// handleError(err, log)
	// log.Debugf("%s", data)
}
