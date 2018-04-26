package main

import (
	"os"
	"os/user"

	"github.com/Sirupsen/logrus"
	"github.com/glnds/masl/internal/masl"
)

var logger = logrus.New()

func main() {

	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}

	// Create the logger file if doesn't exist. And append to it if it already exists.
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

	// Read config file
	conf := masl.GetConfig(logger)

	// Generate a new OneLogin API Token
	masl.GenerateToken(conf, logger)
}
