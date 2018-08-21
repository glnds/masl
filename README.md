# MASL
[![CircleCI](https://circleci.com/gh/glnds/masl.svg?style=svg)](https://circleci.com/gh/glnds/masl)
[![Go Report Card](https://goreportcard.com/badge/github.com/glnds/masl)](https://goreportcard.com/report/github.com/glnds/masl)


![MASL](img/masl.png)


Pronounced [mɑzəl] form the Dutch word 'mazzel', meaning luck. 'masl' is an anagram from the word 'SAML'.
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
Debug = true/false (Set to true for debug logging, default off)
Profile = 'AWS Profile name' (default = masl)
```

## Usage

Just run ```masl``` on your command line. 

Optional command line arguments:
```
  -profile string
        AWS profile name (default "xxxx")
  -version
        prints MASL version
```

Assure the environment variable ```AWS_PROFILE``` is set to **masl** (or the overrided value specified in your config file or command line options).

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
