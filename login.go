package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/user"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("masl")
var format = logging.MustStringFormatter(`%{time:2006-01-02T15:04:05.999999999} %{shortfunc} -  %{level:.5s} %{message}`)

const (
	generateTokenAPI = "auth/oauth2/token"
)

type config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AppID        string
	Subdomain    string
	Username     string
	Debug        bool
}

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

func handleError(err error, logger *logging.Logger) {
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func getJSON(req *http.Request, target interface{}) error {

	resp, err := httpClient.Do(req)
	if err != nil {
		return (err)
	}
	defer resp.Body.Close()
	debugREST(httputil.DumpResponse(resp, true))

	return json.NewDecoder(resp.Body).Decode(target)
}

func main() {

	// Read go-inside.toml for initialization
	var conf config
	if _, err := toml.DecodeFile("./masl.toml", &conf); err != nil {
		log.Error(err.Error())
		panic(err)
	}

	// Get the user home directory
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Initialze a log file
	logFile, err := os.OpenFile(usr.HomeDir+"/masl.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	handleError(err, log)
	loggingBackend := logging.NewLogBackend(logFile, "", 0)
	backendFormatter := logging.NewBackendFormatter(loggingBackend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	if conf.Debug {
		backendLeveled.SetLevel(logging.DEBUG, "")
	} else {
		backendLeveled.SetLevel(logging.INFO, "")
	}
	logging.SetBackend(backendLeveled)

	log.Info("\n\n----- w00t w00t go-inside Initilaized -----\n")
	log.Infof("Current config:\n%#v\n", conf)

	generateToken(conf)
}

// Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
// Generate an access token and refresh token to access onelogin's resource APIs.
func generateToken(conf config) {

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
	handleError(err, log)
	log.Debugf("%s", data)
}
