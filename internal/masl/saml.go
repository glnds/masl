package masl

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
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
type samlAssertionResponse struct {
	Status struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   bool   `json:"error"`
	} `json:"status"`
	Data string `json:"data"`
}

// SAMLAssertionResponseMFA represents the OneLogin SAML Assertion response with MFA required
type samlAssertionResponseMFA struct {
	Status struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   bool   `json:"error"`
	} `json:"status"`
	Data []struct {
		CallbackURL string      `json:"callback_url"`
		Devices     []MFADevice `json:"devices"`
		StateToken  string      `json:"state_token"`
		User        struct {
			Email     string `json:"email"`
			Lastname  string `json:"lastname"`
			Username  string `json:"username"`
			ID        int    `json:"id"`
			Firstname string `json:"firstname"`
		} `json:"user"`
	} `json:"data"`
}

// SAMLAssertionData internal Generic SAMLAssertion response representation
type SAMLAssertionData struct {
	MFARequired bool
	StateToken  string
	Data        string
	Devices     []MFADevice
}

// MFADevice represents an MFA device
type MFADevice struct {
	DeviceID   int    `json:"device_id"`
	DeviceType string `json:"device_type"`
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
	ID                     int
	PrincipalArn           string
	RoleArn                string
	AccountID              string
	AccountName            string
	EnvironmentIndependent bool
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
	// dump, _ := httputil.DumpRequest(req, true)
	//TODO: shame on me, filter passwords from the requests before logging them!
	// log.Debug(string(dump))
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

	// Parse the raw body to determine if MFA is required
	body := httpRequestRaw(url, auth, requestBody, log)
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		log.Fatalln(err)
	}
	status := rawData["status"].(map[string]interface{})
	message := status["message"].(string)
	log.Info(message)

	var samlData SAMLAssertionData
	var samlErr error

	code := int(status["code"].(float64))
	if code == 200 {
		if strings.EqualFold(message, "success") {
			// MFA NOT Required
			log.Info("MFA not required")
			assertionResponse := samlAssertionResponse{}
			json.Unmarshal(body, &assertionResponse)

			samlData = SAMLAssertionData{
				MFARequired: false,
				Data:        assertionResponse.Data,
			}
		} else {
			// MFA token is required
			assertionResponse := samlAssertionResponseMFA{}
			log.WithFields(logrus.Fields{
				"response": assertionResponse,
			}).Debug("Assertionresponse in case of  MFA")

			json.Unmarshal(body, &assertionResponse)

			samlData = SAMLAssertionData{
				MFARequired: true,
				StateToken:  assertionResponse.Data[0].StateToken,
				Devices:     assertionResponse.Data[0].Devices,
				// DeviceID:    assertionResponse.Data[0].Devices[0].DeviceID,
				// DeviceType:  assertionResponse.Data[0].Devices[0].DeviceType,
			}
			// copy(assertionResponse.Data[0].Devices, samlData.Devices)
		}
	} else {
		samlErr = errors.New(message)
	}
	return samlData, samlErr
}

// VerifyMFA Call to https://api.eu.onelogin.com/api/1/saml_assertion/verify_factor
func VerifyMFA(conf Config, log *logrus.Logger, deviceID int, stateToken string, otp string,
	apiToken string) (string, error) {

	url := conf.BaseURL + verifyFactorAPI
	requestBody, err := json.Marshal(VerifyMFARequest{
		AppID:      conf.AppID,
		OtpToken:   otp,
		DeviceID:   strconv.Itoa(deviceID),
		StateToken: stateToken})
	if err != nil {
		log.Fatalln(err)
	}
	auth := "bearer:" + apiToken

	mfaResponse := VerifyMFAResponse{}
	httpRequest(url, auth, requestBody, log, &mfaResponse)

	var samlData string
	var samlErr error

	if mfaResponse.Status.Code == 200 {
		samlData = mfaResponse.Data
	} else {
		samlErr = errors.New(mfaResponse.Status.Message)
	}

	return samlData, samlErr
}

// ParseSAMLAssertion parse the SAMLAssertion response data into a list of SAMLAssertionRoles
func ParseSAMLAssertion(samlAssertion string, accountInfo Accounts, accountFilter []string,
	role string) []*SAMLAssertionRole {

	sDec, _ := b64.StdEncoding.DecodeString(samlAssertion)

	var samlResponse Response
	xml.Unmarshal(sDec, &samlResponse)

	attributes := samlResponse.Assertion.AttributeStatement.Attributes

	roles := []*SAMLAssertionRole{}

	for _, attribute := range attributes {
		for _, value := range attribute.Values {
			if strings.Contains(value.Value, "role") {

				data := strings.Split(value.Value, ",")

				assertionRole := SAMLAssertionRole{
					RoleArn:      data[0],
					PrincipalArn: data[1],
					AccountID:    data[1][13:25],
				}
				assertionRole.AccountName, assertionRole.EnvironmentIndependent =
					SearchAccounts(accountInfo, assertionRole.AccountID)

				// Based on context, are we interested in this role?
				if role == "" || strings.EqualFold(role, assertionRole.RoleArn[31:]) {
					if accountFilter == nil {
						roles = append(roles, &assertionRole)
					} else if Contains(accountFilter, assertionRole.AccountID) {
						roles = append(roles, &assertionRole)
					}
				}
			}
		}
	}
	sort.Sort(RolesByName(roles))
	return roles
}

// AssumeRole assume a role on AWS
func AssumeRole(samlAssertion string, duration int64, role *SAMLAssertionRole,
	log *logrus.Logger) *sts.AssumeRoleWithSAMLOutput {

	session := session.Must(session.NewSession())
	stsClient := sts.New(session)

	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &duration,
		PrincipalArn:    &role.PrincipalArn,
		RoleArn:         &role.RoleArn,
		SAMLAssertion:   &samlAssertion}

	output, err := stsClient.AssumeRoleWithSAML(&input)
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err)
	}
	return output
}

// SetCredentials Apply the STS credentials on the host
func SetCredentials(assertionOutput *sts.AssumeRoleWithSAMLOutput, homeDir string,
	profileName string, legacyToken bool, log *logrus.Logger) {
	filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if filename == "" {
		filename = homeDir + string(os.PathSeparator) + ".aws" +
			string(os.PathSeparator) + "credentials"
	}
	var cfg *ini.File

	ini.PrettyFormat = false

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Info("AWS credentials file does not exists.")
		cfg = ini.Empty()
		log.Info("New AWS credentials config created.")
	} else {
		cfg, err = ini.Load(filename)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("AWS credentials file loaded.")
	}

	sec := cfg.Section(profileName)
	sec.NewKey("aws_access_key_id", *assertionOutput.Credentials.AccessKeyId)
	sec.NewKey("aws_secret_access_key", *assertionOutput.Credentials.SecretAccessKey)
	sec.NewKey("aws_session_token", *assertionOutput.Credentials.SessionToken)
	if legacyToken {
		sec.NewKey("aws_security_token", *assertionOutput.Credentials.SessionToken)
	} else {
		sec.DeleteKey("aws_security_token")
	}
	err := cfg.SaveTo(filename)
	if err != nil {
		log.Fatal(err)
	}
}

// Contains test if an array contains a string
func Contains(anArray []string, aString string) bool {
	for _, value := range anArray {
		if aString == value {
			return true
		}
	}
	return false
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

func httpRequestRaw(url string, auth string, jsonStr []byte, log *logrus.Logger) []byte {

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}
