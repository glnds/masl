BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter
PLATFORMS := windows linux darwin
BINARY := masl
VERSION ?= vlatest

os = $(word 1, $@)

PKGS := $(shell go list ./... | grep -v /vendor)

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

test:
	go test $(PKGS)
.PHONY: test

lint: $(GOMETALINTER)
	gometalinter --vendor --config gometalinter.json  ./...
.PHONY: lint

$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(VERSION)-$(os)-amd64 cmd/masl/masl.go 
.PHONY: $(PLATFORMS)

release: windows linux darwin
.PHONY: release
