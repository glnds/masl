package main

import (
	"os"
	"os/user"

	"bufio"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/glnds/masl/internal/masl"
	"github.com/howeyc/gopass"
)

var logger = logrus.New()

func main() {

	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}

	// Create the logger file if doesn't exist. Append to it if it already exists.
	var filename = "/masl.log"
	file, err := os.OpenFile(usr.HomeDir+filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logger.Formatter = Formatter
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to logger to file, using default stderr")
	}

	logger.Info("--------------- w00t w00t masl for you!?  ---------------")
	logger.SetLevel(logrus.InfoLevel)

	// Read config file
	conf := masl.GetConfig(logger)
	if conf.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// First, generate a new OneLogin API token
	apiToken := masl.GenerateToken(conf, logger)

	// Ask for the user's password
	fmt.Print("OneLogin Password: ")
	password, _ := gopass.GetPasswdMasked()

	// OneLogin SAML assertion API call
	samlAssertionData, err := masl.SAMLAssertion(conf, logger, string(password), apiToken)
	if err != nil {
		fmt.Println(err)
		logger.Fatal(err)
	}

	// Ask for a new otp
	fmt.Print("OneLogin Protect Token: ")
	reader := bufio.NewReader(os.Stdin)
	otp, _ := reader.ReadString('\n')

	// OneLogin Verify MFA API call
	samlAssertion, err := masl.VerifyMFA(conf, logger, samlAssertionData, otp, apiToken)
	if err != nil {
		fmt.Println(err)
		logger.Fatal(err)
	}

	roles := masl.ParseSAMLAssertion(conf, samlAssertion)
	for index, role := range roles {
		role.ID = index + 1
		fmt.Printf("[%2d] > %s:%-15s :: %s\n", role.ID, role.AccountID, role.RoleArn[31:], role.AccountName)
	}
}
