package masl

import (
	"os"
	"os/user"
	"strings"

	"github.com/BurntSushi/toml"
)

// Accounts represents the accounts section of the masl config file
type Accounts []struct {
	ID                     string `toml:"ID"`
	Name                   string `toml:"Name"`
	EnvironmentIndependent bool   `toml:"EnvironmentIndependent"`
}

// Config represents the masl config file
type Config struct {
	BaseURL         string `toml:"BaseURL"`
	ClientID        string `toml:"ClientID"`
	ClientSecret    string `toml:"ClientSecret"`
	AppID           string `toml:"AppID"`
	Subdomain       string `toml:"Subdomain"`
	Username        string `toml:"Username"`
	Duration        int    `toml:"Duration"`
	Profile         string `toml:"Profile"`
	DefaultRole     string `toml:"DefaultRole"`
	LegacyToken     bool   `toml:"LegacyToken"`
	Debug           bool   `toml:"Debug"`
	DefaulMFADevice string `toml:"DefaulMFADevice"`
	Environments    []struct {
		Name     string   `toml:"Name"`
		Accounts []string `toml:"Accounts"`
	} `toml:"Environments"`
	Accounts Accounts `toml:"Accounts"`
}

var logger = GetInstance()

// GetConfig reads the .masl/config.toml configuration file for initialization.
func GetConfig() Config {

	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Read .masl/config.toml config file for initialization
	conf := Config{Profile: "masl", LegacyToken: false, Debug: false, Duration: 3600} // Set default values
	if _, err := toml.DecodeFile(usr.HomeDir+string(os.PathSeparator)+".masl"+string(os.PathSeparator)+"config.toml", &conf); err != nil {
		logger.Fatal(err.Error())
	}

	return conf
}

// SearchAccounts search an account name for a given acount id
func SearchAccounts(accountInfo Accounts, accountID string) (string, bool) {

	for _, account := range accountInfo {
		if account.ID == accountID {
			return account.Name, account.EnvironmentIndependent
		}
	}
	return "untitled", false
}

// GetAccountID get the account id for a given acount name (alias)
func GetAccountID(conf Config, name string) string {
	var id string
	for _, account := range conf.Accounts {
		if strings.EqualFold(account.Name, name) {
			id = account.ID
		}
	}
	return id
}

// GetAccountsForEnvironment search an environment's detail for a given environment name
func GetAccountsForEnvironment(conf Config, environment string) []string {
	var accounts []string
	for _, env := range conf.Environments {
		if strings.EqualFold(env.Name, environment) {
			accounts = append(accounts, env.Accounts...)
			break
		}
	}
	for _, account := range conf.Accounts {
		if account.EnvironmentIndependent {
			accounts = append(accounts, account.ID)
		}
	}
	return accounts
}
