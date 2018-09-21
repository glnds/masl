# MASL
[![CircleCI](https://circleci.com/gh/glnds/masl.svg?style=svg)](https://circleci.com/gh/glnds/masl)
[![Go Report Card](https://goreportcard.com/badge/github.com/glnds/masl)](https://goreportcard.com/report/github.com/glnds/masl)


![MASL](img/masl.png)


Pronounced [mɑzəl] form the Dutch word 'mazzel', meaning luck. 'masl' is also an anagram from the word 'SAML'.
This tool allows you to use [onelogin](https://www.onelogin.com/) to assume an AWS role through SAML authentication.

## Getting Started

### Installation

Just download the latest release under https://github.com/glnds/masl/releases.

### Configuration

All configuration is done using a ```masl.toml``` file in your user's home directory.
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
Debug = true/false (Set to true for debug logging, default off)
Profile = 'Value for environment variable AWS_PROFILE' (default = 'masl')
```

If specifying a custom duration assure this duration is allowed on the AWS role itself as well. 
See: [Enable Federated API Access to your AWS Resources for up to 12 hours Using IAM Roles](https://aws.amazon.com/blogs/security/enable-federated-api-access-to-your-aws-resources-for-up-to-12-hours-using-iam-roles/)

#### Multi-Account management
One of the main drivers to develop another Onelogin CLI authenticator was to ease the management of multiple AWS accounts. Most of the tools currently lack those features and that makes switching AWS accounts bothersome. For this purpose ```masl.toml``` supports the following features:

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

Assure the environment variable ```AWS_PROFILE``` is set to **masl** (or the overrided value specified in ```masl.toml``` or the ```-profile``` command line optiont).

## Development

### Dependency management
Dependency management is done with ```dep```: https://github.com/golang/dep

After ```git clone``` run ```dep ensure``` to make sure the project's dependencies are in sync.

Please see [Daily Dep](https://golang.github.io/dep/docs/daily-dep.html) for more information about common ```dep``` commands.

### Makefile
This project includes a ```makefile`` to make your life easy.
- ```make clean```: clean up your workspace
- ```make build```: build this project
- ```make lint```: run [gometalinter](https://github.com/alecthomas/gometalinter)




## Running the tests

TODO: Explain how to run the automated tests for this system




## Built With

* [Snyk](https://snyk.io/) - Continuously vulnerabilities scanning
* [dep](https://golang.github.io/dep/) - Dependency Management for Go

### Logging

A log file ```masl.log``` is created and added on your user's home directory. The default log level is 'INFO'. For debug logging set ```Debug = true``` in ```masl.toml```.

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :Do us.

## Versioning

[SemVer](http://semver.org/) is used for versioning. For the versions available, see the [tags on this repository](https://github.com/glnds/masl/tags). 

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
