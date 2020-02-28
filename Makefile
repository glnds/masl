SHELL := /bin/bash

BIN_DIR := $(GOPATH)/bin
PLATFORMS := windows linux darwin
BINARY := masl

# These will be provided to the target
VERSION := 2.0.2
BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.version=$(VERSION) -X=main.build=$(BUILD)"

os = $(word 1, $@)

PKGS := $(shell go list ./... | grep -v /vendor)

clean:
	go clean
	rm -f masl
	rm -f masl.exe
	rm -rf release/
.PHONY: clean

build:
	go build $(LDFLAGS) cmd/masl/masl.go
.PHONY: build

test:
	go test $(PKGS)
.PHONY: test

lint:
	golangci-lint run
.PHONY: lint

# $(PLATFORMS):
# 	mkdir -p release
# 	GOOS=$(os) GOARCH=amd64 go build $(LDFLAGS) -o release/$(BINARY)-v$(VERSION)-$(os)-amd64 cmd/masl/masl.go
# .PHONY: $(PLATFORMS)

# # run "make release -j3
# release: windows linux darwin
# .PHONY: release

install:
	@go install $(LDFLAGS) cmd/masl/masl.go

# LDFLAGS are parsed by goreleaser
# Default is `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser`
release:
	@goreleaser $(GORELEASER_ARGS)
.PHONY: release

snapshot: GORELEASER_ARGS= --rm-dist --snapshot
snapshot: release
.PHONY: snapshot

todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=dist \
		--exclude-dir=Attic \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E 'TODO:.*' .
