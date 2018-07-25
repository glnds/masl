package masl

import (
	"log"
	"os/user"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
)

// MaslConfig represents the masl config file
type MaslConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	AppID        string
	Subdomain    string
	Username     string
	Debug        bool
	Accounts     []struct {
		ID   string
		Name string
	}
}

// GetMaslConfig reads the masl.toml configuration file for initialization.
func GetMaslConfig(logger *logrus.Logger) MaslConfig {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Read masl.toml config file for initialization
	var conf MaslConfig
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
		"#accounts":    len(conf.Accounts),
	}).Info("Config settings")

	return conf
}

// SearchAccounts search an account name for a given acount id
func SearchAccounts(conf MaslConfig, accountID string) string {

	for _, account := range conf.Accounts {
		if account.ID == accountID {
			return account.Name
		}
	}
	return ""
}
