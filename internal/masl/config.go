package masl

import (
	"log"
	"os/user"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
)

// Config represents the masl config file
type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AppID        string
	Subdomain    string
	Username     string
	Debug        bool
}

// GetConfig reads the masl.toml configuration file for initialization.
func GetConfig(logger *logrus.Logger) Config {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Read masl.toml config file for initialization
	var conf Config
	if _, err := toml.DecodeFile(usr.HomeDir+"/masl.toml", &conf); err != nil {
		log.Fatal(err.Error())
	}

	logger.WithFields(logrus.Fields{
		"baseURL":      conf.BaseURL,
		"clientID":     conf.ClientID,
		"clientSecret": conf.ClientSecret,
		"appID":        conf.AppID,
		"subdomain":    conf.Subdomain,
		"username":     conf.Username,
		"debug":        conf.Debug,
	}).Info("Config settings")

	return conf
}
