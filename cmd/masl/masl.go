package main

import (
	"os"
	"os/user"

	"github.com/Sirupsen/logrus"
	"github.com/glnds/masl/internal/masl"
)

var log = logrus.New()

func main() {

	conf := masl.GetConfig()

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Create the log file if doesn't exist. And append to it if it already exists.
	var filename = "/masl2.log"
	file, err := os.OpenFile(usr.HomeDir+filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.Formatter = Formatter
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	log.Info("--------------- w00t w00t go-inside Initilaized ---------------")
	log.Infof("Current config:\n%#v\n", conf)

	masl.Login(conf, log)
}
