package masl

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type OneloginClient struct {
	c            *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	apiToken     string
	username     string
	appID        string
	subdomain    string
}

type status struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// APITokenResponse represents a OneLogin Generate API Token response
type APITokenResponse struct {
	Status status `json:"status"`
	Data   []struct {
		AccessToken  string    `json:"access_token"`
		CreatedAt    time.Time `json:"created_at"`
		ExpiresIn    int       `json:"expires_in"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		AccountID    int       `json:"account_id"`
	} `json:"data"`
}

// ErrorResponse represents a OneLogin Error response
type ErrorResponse struct {
	Status status `json:"status"`
}

func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("%d (%s) API error: %s", err.Status.Code, err.Status.Type, err.Status.Message)
}

func NewOneloginClient(config Config) *OneloginClient {
	client := &http.Client{Timeout: 10 * time.Second}

	return &OneloginClient{
		c:            client,
		baseURL:      config.BaseURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		username:     config.Username,
		appID:        config.AppID,
		subdomain:    config.Subdomain,
	}
}

// InitApiToken Call to https://developers.onelogin.com/api-docs/1/oauth20-tokens/generate-tokens
func (client *OneloginClient) InitApiToken() error {

	url := client.baseURL + "auth/oauth2/token"
	requestBody := []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	auth := "client_id:" + client.clientID + ",client_secret:" + client.clientSecret
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.c.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		var apiToken APITokenResponse
		if err := json.NewDecoder(res.Body).Decode(&apiToken); err != nil {
			return err
		}
		client.apiToken = apiToken.Data[0].AccessToken
		return nil
	case 400, 401, 404:
		var errRes ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return err
		}
		return &errRes
	default:
		// handle unexpected status codes
		return fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
}

// SAMLAssertion Call to https://api.eu.onelogin.com/api/1/saml_assertion
func (client *OneloginClient) SAMLAssertion(log *logrus.Logger, password string) (SAMLAssertionData, error) {

	url := client.baseURL + "api/1/saml_assertion"
	requestBody, err := json.Marshal(SAMLAssertionRequest{
		UsernameOrEmail: client.username,
		Password:        password,
		AppID:           client.appID,
		Subdomain:       client.subdomain})
	if err != nil {
		return SAMLAssertionData{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return SAMLAssertionData{}, err
	}

	req.Header.Set("Authorization", "bearer:"+client.apiToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.c.Do(req)
	if err != nil {
		return SAMLAssertionData{}, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return SAMLAssertionData{}, err
	}

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
			if err := json.Unmarshal(body, &assertionResponse); err != nil {
				log.Fatalln(err)
			}

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

			if err := json.Unmarshal(body, &assertionResponse); err != nil {
				log.Fatalln(err)
			}

			samlData = SAMLAssertionData{
				MFARequired: true,
				StateToken:  assertionResponse.Data[0].StateToken,
				Devices:     assertionResponse.Data[0].Devices,
			}
		}
	} else {
		// TODO: enhance
		samlErr = errors.New(message)
	}
	return samlData, samlErr
}

// VerifyMFA Call to https://api.eu.onelogin.com/api/1/saml_assertion/verify_factor
func (client *OneloginClient) VerifyMFA(log *logrus.Logger, deviceID int, stateToken string,
	otp string) (string, error) {

	url := client.baseURL + "api/1/saml_assertion/verify_factor"
	requestBody, err := json.Marshal(VerifyMFARequest{
		AppID:      client.appID,
		OtpToken:   otp,
		DeviceID:   strconv.Itoa(deviceID),
		StateToken: stateToken})
	if err != nil {
		log.Fatalln(err)
	}
	auth := "bearer:" + client.apiToken

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

// RolesByName roles sorted by account name
type RolesByName []*SAMLAssertionRole

func (byName RolesByName) Len() int      { return len(byName) }
func (byName RolesByName) Swap(i, j int) { byName[i], byName[j] = byName[j], byName[i] }
func (byName RolesByName) Less(i, j int) bool {
	return strings.Compare(byName[i].AccountName, byName[j].AccountName) == -1
}

// ParseSAMLAssertion parse the SAMLAssertion response data into a list of SAMLAssertionRoles
func ParseSAMLAssertion(samlAssertion string, accountInfo Accounts, log *logrus.Logger,
	accountFilter []string, role string) []*SAMLAssertionRole {

	sDec, _ := b64.StdEncoding.DecodeString(samlAssertion)

	var samlResponse Response
	if err := xml.Unmarshal(sDec, &samlResponse); err != nil {
		log.Fatalln(err)
	}

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
