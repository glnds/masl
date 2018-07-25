package masl

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/ini.v1"
)

/* #nosec */
const (
	generateTokenAPI = "auth/oauth2/token"
	samlAssertionAPI = "api/1/saml_assertion"
	verifyFactorAPI  = "api/1/saml_assertion/verify_factor"
)

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

// SAMLAssertionRequest represents the OneLogin SAML Assertion request
type SAMLAssertionRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	AppID           string `json:"app_id"`
	Subdomain       string `json:"subdomain"`
}

// SAMLAssertionResponse represents the OneLogin SAML Assertion response
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

// SAMLAssertionData internal SAMLAssertionResponse representation
type SAMLAssertionData struct {
	StateToken string
	DeviceID   int
}

// VerifyMFARequest represents the OneLogin Verify MFA request
type VerifyMFARequest struct {
	AppID      string `json:"app_id"`
	OtpToken   string `json:"otp_token"`
	DeviceID   string `json:"device_id"`
	StateToken string `json:"state_token"`
}

// VerifyMFAResponse represents the OneLogin Verify MFA response
type VerifyMFAResponse struct {
	Status struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   bool   `json:"error"`
	} `json:"status"`
	Data string `json:"data"`
}

// SAMLAssertionRole represents a Role which could be assumed on AWS
type SAMLAssertionRole struct {
	ID           int
	PrincipalArn string
	RoleArn      string
	AccountID    string
	AccountName  string
}

// RolesByName roles sorted by account name
type RolesByName []*SAMLAssertionRole

func (byName RolesByName) Len() int      { return len(byName) }
func (byName RolesByName) Swap(i, j int) { byName[i], byName[j] = byName[j], byName[i] }
func (byName RolesByName) Less(i, j int) bool {
	return strings.Compare(byName[i].AccountName, byName[j].AccountName) == -1
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

// GenerateToken Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
func GenerateToken(conf Config, log *logrus.Logger) string {

	url := conf.BaseURL + generateTokenAPI
	requestBody := []byte(`{"grant_type":"client_credentials"}`)
	auth := "client_id:" + conf.ClientID + ",client_secret:" + conf.ClientSecret

	apiToken := APITokenResponse{}
	httpRequest(url, auth, requestBody, log, &apiToken)

	//TODO: being a bit optimistic here ;)
	return apiToken.Data[0].AccessToken
}

// SAMLAssertion Call to https://api.eu.onelogin.com/api/1/saml_assertion
func SAMLAssertion(conf Config, log *logrus.Logger, password string, apiToken string) (SAMLAssertionData, error) {

	url := conf.BaseURL + samlAssertionAPI
	requestBody, err := json.Marshal(SAMLAssertionRequest{
		UsernameOrEmail: conf.Username,
		Password:        password,
		AppID:           conf.AppID,
		Subdomain:       conf.Subdomain})
	if err != nil {
		log.Fatalln(err)
	}
	auth := "bearer:" + apiToken

	samlAssertion := SAMLAssertionResponse{}
	httpRequest(url, auth, requestBody, log, &samlAssertion)

	var data SAMLAssertionData
	if samlAssertion.Status.Code == 200 {
		//TODO: MFA-less authentication is not yet implemented
		data = SAMLAssertionData{
			samlAssertion.Data[0].StateToken,
			samlAssertion.Data[0].Devices[0].DeviceID}
		return data, nil
	}
	return data, errors.New(samlAssertion.Status.Message)
}

// VerifyMFA Call to https://api.eu.onelogin.com/api/1/saml_assertion/verify_factor
func VerifyMFA(conf Config, log *logrus.Logger, data SAMLAssertionData, otp string,
	apiToken string) (string, error) {

	url := conf.BaseURL + verifyFactorAPI
	requestBody, err := json.Marshal(VerifyMFARequest{
		AppID:      conf.AppID,
		OtpToken:   otp,
		DeviceID:   strconv.Itoa(data.DeviceID),
		StateToken: data.StateToken})
	if err != nil {
		log.Fatalln(err)
	}
	auth := "bearer:" + apiToken

	mfaResponse := VerifyMFAResponse{}
	httpRequest(url, auth, requestBody, log, &mfaResponse)

	if mfaResponse.Status.Code == 200 {
		return mfaResponse.Data, nil
	}

	return "", errors.New(mfaResponse.Status.Message)
}

// ParseSAMLAssertion parse the SAMLAssertion response data into a list of SAMLAssertionRoles
func ParseSAMLAssertion(conf Config, samlAssertion string) []*SAMLAssertionRole {

	sDec, _ := b64.StdEncoding.DecodeString(samlAssertion)
	fmt.Printf("%v\n", string(sDec[:]))

	var samlResponse Response
	xml.Unmarshal(sDec, &samlResponse)

	attributes := samlResponse.Assertion.AttributeStatement.Attributes

	roles := []*SAMLAssertionRole{}

	for i := 0; i < len(attributes); i++ {
		values := attributes[i].Values
		for j := 0; j < len(values); j++ {
			if strings.Contains(values[j].Value, "role") {

				data := strings.Split(values[j].Value, ",")

				role := new(SAMLAssertionRole)
				role.RoleArn = data[0]
				role.PrincipalArn = data[1]
				role.AccountID = data[1][13:25]
				role.AccountName = SearchAccounts(conf, role.AccountID)

				roles = append(roles, role)
			}
		}
	}
	sort.Sort(RolesByName(roles))
	return roles
}

func httpRequest(url string, auth string, jsonStr []byte, log *logrus.Logger, target interface{}) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	logRequest(log, req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	logResponse(log, resp)

	json.NewDecoder(resp.Body).Decode(target)
}

// AssumeRole assume a role on AWS
func AssumeRole(samlAssertion string, role *SAMLAssertionRole, log *logrus.Logger) *sts.AssumeRoleWithSAMLOutput {
	session := session.Must(session.NewSession())
	stsClient := sts.New(session)

	duration := int64(28800)
	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &duration,
		PrincipalArn:    &role.PrincipalArn,
		RoleArn:         &role.RoleArn,
		SAMLAssertion:   &samlAssertion}

	output, err := stsClient.AssumeRoleWithSAML(&input)
	if err != nil {
		log.Fatal(err)
	}
	return output
}

// SetCredentials Apply the STS credentials on the host
func SetCredentials(assertionOutput *sts.AssumeRoleWithSAMLOutput, homeDir string, log *logrus.Logger) {
	filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if filename == "" {
		filename = homeDir + "/.aws/credentials"
	}
	cfg, err := ini.Load(filename)
	if err != nil {
		log.Fatal(err)
	}
	sec := cfg.Section("masl")
	sec.NewKey("aws_access_key_id", *assertionOutput.Credentials.AccessKeyId)
	sec.NewKey("aws_secret_access_key", *assertionOutput.Credentials.SecretAccessKey)
	sec.NewKey("aws_session_token", *assertionOutput.Credentials.SessionToken)
	err = cfg.SaveTo(homeDir + "/.aws/credentials")
	if err != nil {
		log.Fatal(err)
	}
}
