package main

import (
	"os"
	"os/user"

	"github.com/Sirupsen/logrus"
	"github.com/glnds/masl/internal/masl"
	"fmt"
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

	// As for the user's password
	fmt.Print("OneLogin Password: ")
	password, err := gopass.GetPasswdMasked()
	if err != nil {
		logger.Fatal(err)
	}
	// Generate a new OneLogin API APITokenResponse
	apiToken := masl.GenerateToken(conf, logger)
	masl.SAMLAssertion(conf, logger, string(password), apiToken)
}
