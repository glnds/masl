package masl

import (
	"log"
	"os/user"

	"github.com/BurntSushi/toml"
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
func GetConfig() Config {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Read masl.toml confif file for initialization
	var conf Config
	if _, err := toml.DecodeFile(usr.HomeDir+"/masl.toml", &conf); err != nil {
		log.Fatal(err.Error())
	}

	return conf
}
