[![Release](https://img.shields.io/github/release/glnds/masl.svg?style=flat)](https://github.com/glnds/masl/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/glnds/masl)](https://goreportcard.com/report/github.com/glnds/masl)
[![CircleCI](https://circleci.com/gh/glnds/masl.svg?style=svg)](https://circleci.com/gh/glnds/masl)


# MASL

![MASL](img/masl.png)


Pronounced [mɑzəl] form the Dutch word 'mazzel', meaning luck. 'masl' is also an anagram from the word 'SAML'.
This tool allows you to use [onelogin](https://www.onelogin.com/) to assume an AWS role through SAML authentication.

## Getting Started

### Installation

Just download the latest release under https://github.com/glnds/masl/releases. Don't forget to make it executable and link it from somewhere in your $PATH on UNIX-like platforms.

#### OS X
- Open Masl.dmg
- Move the file **masl-vx.x.x-darwin-amd64** to /usr/local/bin and rename it to **masl**.
- done :)

#### Windows
- rename **masl-vx.x.x-windows-amd64** to **masl.exe** and execute to install.


### Configuration

All configuration is done using a `.masl/config.toml` file in your user's home directory.
An example toml config file is included: [masl-example.toml](https://github.com/glnds/masl/blob/master/masl-example.toml).
Copy `masl-example.toml` and rename it to `.masl/config.toml`. Adjust the values to reflect your environment.


The minimal configuration should look like this:
```

BaseURL = 'https://api.eu.onelogin.com/'
ClientID = 'onelogin client id'
ClientSecret = 'onelogin client secret'
AppID = 'onelogin app id'
Subdomain = 'subdomain of the onelogin user'
Username = 'onelogin username or email'
```

Optional settings:
```
Duration = 'Assume role maximum session duration' (default 3600)
LegacyToken = true/false (configures legacy aws_security_token (for Boto support))
Debug = true/false (Set to true for debug logging, default off)
Profile = 'Value for environment variable AWS_PROFILE' (default = 'masl')
DefaulMFADevice = 'name of your default MFA device (for example 'Yubico YubiKey')'
```

If specifying a custom duration assure this duration is allowed on the AWS role itself as well. 
See: [Enable Federated API Access to your AWS Resources for up to 12 hours Using IAM Roles](https://aws.amazon.com/blogs/security/enable-federated-api-access-to-your-aws-resources-for-up-to-12-hours-using-iam-roles/)

#### Multi-Account management
One of the main drivers to develop another Onelogin CLI authenticator was to ease the management of multiple AWS accounts. Most of the tools currently lack those features and that makes switching AWS accounts bothersome. For this purpose ```.masl/config.toml``` supports the following features:

##### Account naming
You can provide account names (aliases) for all accounts you have access to:
```
...
[[Accounts]]
ID = '1234567890'
Name = 'account-x'

[[Accounts]]
ID = '1122334455'
Name = 'account-y'

[[Accounts]]
ID = '0987654321'
Name = 'account-z'
...
```

##### Environments containing account subsets
If your account list grows too big it is often handy to limit the list to your current work context. This can be achieved by defining environments:

```
...
[[Environments]]
Name = 'governance'
Accounts = ['1234567890', '1122334455']
...
```

Furthermore accounts can be marked as 'Environment Independent`, in that case they will show up in all your environments.

```
...
[[Accounts]]
ID = '1234567890'
Name = 'base-account'
EnvironmentIndependent = true
...
````

usage: ```masl -env [environment_name]```


## Usage

Just run ```masl``` on your command line. 

Optional command line arguments:
```
  -account string
        AWS Account ID or name
  -env string
        Work environment
  -legacy-token
        configures legacy aws_security_token (for Boto support)
  -profile string
        AWS profile name (default "masl")
  -role string
        AWS role name
  -version
        prints MASL version
```

Assure the environment variable ```AWS_PROFILE``` is set to **masl** (or the overrided value specified in ```.masl/config.toml``` or the ```-profile``` command line option).

### Non-interactive usage

If you use command line tools to manage your passwords and generate otp tokens then you can pipe credentials to masl. Masl expects three lines. The first line should be your password, the second line the otp token and the third line the role you want to select.

For example if you use [pass](https://www.passwordstore.org/) to manage your passwords and [totp-cli](https://github.com/WhyNotHugo/totp-cli) to generate tokens, then you can write a script like this:
```
#!/usr/bin/env bash

echo $(pass <the-service>)
echo $(totp <the-service>)
echo 1

```

And pipe the output of this `masl`.

## Development

### Makefile
This project includes a ```makefile`` to make your life easy.
- ```make clean```: clean up your workspace
- ```make build```: build this project
- ```make lint```: run `golangci-lint run`


## Running the tests

TODO: Explain how to run the automated tests for this system


## Built With

* [Snyk](https://snyk.io/) - Continuously vulnerabilities scanning
* [Go Modules](https://github.com/golang/go/wiki/Modules)

### Logging

A log file ```masl.log``` is created and added on your user's home directory. The default log level is 'INFO'. For debug logging set ```Debug = true``` in ```.masl/config.toml```.

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :Do us.

## Versioning

[SemVer](http://semver.org/) is used for versioning. For the versions available, see the [tags on this repository](https://github.com/glnds/masl/tags). 


## FAQ
### My login is successful but I'm unable to access my AWS account?
The AWS CLI uses ```default``` as default for the variable ```AWS_PROFILE```. MASL uses ```masl``` as default AWS profile name to store the AWS credentials (as a safety net for not overriding your default settings). 

To fix this do one of the following:
- set the value of ```AWS_PROFILE``` to ```masl```
- in your ```.masl/config.toml``` add the line ```Profile = 'default'```
- start masl with the ```-profile default``` option

### I have multiple MFA devices defined, is it possible to set one of them as default?
yes in your ```.masl/config.toml``` set a value for the variable ```DefaulMFADevice```

## License
This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
