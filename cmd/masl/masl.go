package main

import (
	"os"
	"os/user"
	"syscall"

	"bufio"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/glnds/masl/internal/masl"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

var logger = logrus.New()

var version, build string

// Flags represents the command line flags
type Flags struct {
	Version     bool
	LegacyToken bool
	Profile     string
	Env         string
	Account     string
	Role        string
}

func main() {
	usr, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		// Ask for the user's password
		fmt.Print("OneLogin Password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin)) // nolint
		password = string(bytePassword)
	}

	logger, file := InitializeLogger(usr)
	defer file.Close()
	conf := LoadConfig(logger)
	flags := parseCliFlags(conf)
	DoMasl(conf, flags, logger, password, usr)
}

func DoMasl(conf masl.Config, flags Flags, logger *logrus.Logger, password string, usr *user.User) {
	accountFilter := initAccountFilter(conf, flags, logger)
	// Generate a new OneLogin API token
	apiToken := masl.GenerateToken(conf, logger)

	// OneLogin SAML assertion API call
	samlAssertionData, err := masl.SAMLAssertion(conf, logger, password, apiToken)
	if err != nil {
		fmt.Printf("\n%s\n", err)
		logger.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	samlData := ReadSamlData(samlAssertionData, conf, reader, apiToken)

	// Print all SAMLAssertion Roles
	roles := masl.ParseSAMLAssertion(samlData, conf.Accounts, logger, accountFilter, flags.Role)
	if len(roles) == 0 {
		fmt.Println("No  masl for you! You don't have permissions to any account!")
		os.Exit(0)
	}
	role := selectRole(roles)
	awsAuthenticate(samlData, conf, role, usr.HomeDir, flags)
}

func LoadConfig(logger *logrus.Logger) masl.Config {
	conf := masl.GetConfig(logger)
	if conf.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	return conf
}

func parseCliFlags(conf masl.Config) Flags {
	flags := parseFlags(conf)
	logger.WithFields(logrus.Fields{
		"flags": flags,
	}).Info("Parsed the commandline flags")
	return flags
}

func InitializeLogger(usr *user.User) (*logrus.Logger, *os.File) {
	// Create the logger file if doesn't exist. Append to it if it already exists.
	var filename = "masl.log"
	file, err := os.OpenFile(usr.HomeDir+string(os.PathSeparator)+".masl"+string(os.PathSeparator)+filename,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logger.Formatter = Formatter
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.Info("------------------ w00t w00t masl for you!?  ------------------")
	logger.SetLevel(logrus.InfoLevel)
	return logger, file
}

func ReadSamlData(samlAssertionData masl.SAMLAssertionData, conf masl.Config, reader *bufio.Reader, apiToken string) string {
	var samlData string
	var err error
	if samlAssertionData.MFARequired {
		fmt.Print("\n")
		device := selectMFADevice(samlAssertionData.Devices, conf.DefaulMFADevice)
		otp := os.Getenv("OTP")
		if otp == "" {
			// Ask for a new otp
			if strings.Contains(strings.ToLower(device.DeviceType), "yubikey") {
				fmt.Printf("Enter your YubiKey security code: ")
			} else {
				fmt.Printf("Enter your %s one-time password: ", device.DeviceType)
			}
			otp, _ = reader.ReadString('\n')
		}
		samlData, err = masl.VerifyMFA(conf, logger, device.DeviceID, samlAssertionData.StateToken,
			otp, apiToken)
		// OneLogin Verify MFA API call
		if err != nil {
			fmt.Println(err)
			logger.Fatal(err)
		}
	} else {
		fmt.Println()
		samlData = samlAssertionData.Data
	}
	return samlData
}

func awsAuthenticate(samlData string, conf masl.Config, role *masl.SAMLAssertionRole,
	homeDir string, flags Flags) {
	assertionOutput := masl.AssumeRole(samlData, int64(conf.Duration), role, logger)
	masl.SetCredentials(assertionOutput, homeDir, flags.Profile, flags.LegacyToken, logger)    //profile
	masl.SetCredentials(assertionOutput, homeDir, role.AccountName, flags.LegacyToken, logger) // account name

	logger.Info("w00t w00t masl for you!, Successfully authenticated.")

	fmt.Println("\nw00t w00t masl for you!")
	fmt.Printf("Assumed User: %v\n", *assertionOutput.AssumedRoleUser.Arn)
	fmt.Printf("In account: %v [%v]\n", role.AccountID, role.AccountName)
	fmt.Printf("Token will expire on: %v\n", *assertionOutput.Credentials.Expiration)
	awsProfile := os.Getenv("AWS_PROFILE")
	if awsProfile == "" {
		awsProfile = "default"
	}
	if flags.Profile != awsProfile {
		fmt.Printf("\033[1;33m[WARNING] Your AWS credentials were stored under profile ")
		fmt.Printf("'%s' but your AWS_PROFILE is set to '%s'!\n", flags.Profile, awsProfile)
		fmt.Print("Please read the FAQ in the README (https://github.com/glnds/masl) ")
		fmt.Println("in order to fix this.\033[0m")
	} else {
		fmt.Printf("\033[1;32mUsing AWS Profile(s): '%v' & '%v'\033[0m\n", flags.Profile, role.AccountName)
	}
}

func parseFlags(conf masl.Config) Flags {
	flags := new(Flags)

	flag.BoolVar(&flags.Version, "version", false, "prints MASL version")
	flag.BoolVar(&flags.LegacyToken, "legacy-token", conf.LegacyToken,
		"configures legacy aws_security_token (for Boto support)")
	flag.StringVar(&flags.Profile, "profile", conf.Profile, "AWS profile name")
	flag.StringVar(&flags.Env, "env", "", "Work environment")
	flag.StringVar(&flags.Account, "account", "", "AWS Account ID or name")
	flag.StringVar(&flags.Role, "role", "", "AWS role name")

	flag.Parse()

	if flags.Version {
		fmt.Printf("masl version: %s, build: %s\n", version, build)
		os.Exit(0)
	}
	return *flags
}

func initAccountFilter(conf masl.Config, flags Flags, log *logrus.Logger) []string {

	var accountFilter []string
	if flags.Account != "" {
		account := masl.GetAccountID(conf, flags.Account)
		if account != "" {
			accountFilter = append(accountFilter, account)
		} else {
			accountFilter = append(accountFilter, flags.Account)
		}
	} else if flags.Env != "" {
		accountFilter = append(accountFilter, masl.GetAccountsForEnvironment(conf, flags.Env)...)
	}
	log.WithFields(logrus.Fields{
		"accountFilter": accountFilter,
	}).Info("Initialized the account filter")

	return accountFilter
}

func selectRole(roles []*masl.SAMLAssertionRole) *masl.SAMLAssertionRole {
	if len(roles) == 1 {
		return roles[0]
	}

	for index, role := range roles {
		role.ID = index + 1
		fmt.Printf("[%2d] > %s:%-15s :: %s\n", role.ID, role.AccountID, role.RoleArn[31:], role.AccountName)
	}

	// Choose a role
	fmt.Print("Enter a role number:")
	reader := bufio.NewReader(os.Stdin)
	roleNumber, _ := reader.ReadString('\n')
	roleNumber = strings.TrimRight(roleNumber, "\r\n")
	index, err := strconv.Atoi(roleNumber)
	if err != nil {
		fmt.Println(err)
		logger.Fatal(err)
	}
	return roles[index-1]
}

func selectMFADevice(devices []masl.MFADevice, defaultMFADevice string) masl.MFADevice {
	if len(devices) == 1 {
		return devices[0]
	}

	if defaultMFADevice != "" {
		// Try to select the default MFA device
		for _, device := range devices {
			if strings.EqualFold(device.DeviceType, defaultMFADevice) {
				fmt.Printf("Picked your default defined MFA device.\n")
				return device
			}
		}
		fmt.Printf("No MFA device match found for your default defined MFA Device: [%s].\n",
			defaultMFADevice)
	}
	// Manually select an MFA device
	for index, device := range devices {
		fmt.Printf("[%2d] > %s\n", index+1, device.DeviceType)
	}
	fmt.Print("Enter the MFA device number:")
	reader := bufio.NewReader(os.Stdin)
	deviceNumber, _ := reader.ReadString('\n')
	deviceNumber = strings.TrimRight(deviceNumber, "\r\n")
	index, err := strconv.Atoi(deviceNumber)
	if err != nil {
		fmt.Println(err)
		logger.Fatal(err)
	}
	return devices[index-1]
}
