package masl

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

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

	var cfg *ini.File
	ini.PrettyFormat = false

	filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if filename == "" {
		path := homeDir + string(os.PathSeparator) + ".aws"
		filename = path + string(os.PathSeparator) + "credentials"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.Mkdir(path, 0755); err != nil {
				log.Fatalln(err)
			}
			log.Info(".aws directory created.")
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			emptyFile, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			emptyFile.Close()
			if err := os.Chmod(filename, 0600); err != nil {
				log.Fatal(err)
			}
			log.Info("AWS credentials file created.")
		}
	}

	var err error
	cfg, err = ini.Load(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("AWS credentials file loaded.")

	sec := cfg.Section(profileName)
	if _, err := sec.NewKey("aws_access_key_id", *assertionOutput.Credentials.AccessKeyId); err != nil {
		log.Fatalln(err)
	}
	if _, err := sec.NewKey("aws_secret_access_key", *assertionOutput.Credentials.SecretAccessKey); err != nil {
		log.Fatalln(err)
	}
	if _, err := sec.NewKey("aws_session_token", *assertionOutput.Credentials.SessionToken); err != nil {
		log.Fatalln(err)
	}
	if legacyToken {
		if _, err := sec.NewKey("aws_security_token", *assertionOutput.Credentials.SessionToken); err != nil {
			log.Fatalln(err)
		}
	} else {
		sec.DeleteKey("aws_security_token")
	}
	err = cfg.SaveTo(filename)
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

// func httpRequest(url string, auth string, jsonStr []byte, log *logrus.Logger, target interface{}) {

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
// 	if err != nil {
// 		panic(err)
// 	}
// 	req.Header.Set("Authorization", auth)
// 	req.Header.Set("Content-Type", "application/json")
// 	// logRequest(log, req)

// 	resp, err := httpClient.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	logResponse(log, resp)

// 	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
// 		log.Fatalln(err)
// 	}
// }

// func httpRequestRaw(url string, auth string, jsonStr []byte, log *logrus.Logger) []byte {

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
// 	if err != nil {
// 		panic(err)
// 	}
// 	req.Header.Set("Authorization", auth)
// 	req.Header.Set("Content-Type", "application/json")
// 	// logRequest(log, req)

// 	resp, err := httpClient.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	logResponse(log, resp)

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return body
// }
