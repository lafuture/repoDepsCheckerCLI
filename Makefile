.PHONY: build clean

VERSION ?= dev
BINARY := go-repo-deps-checker

build:
	go build -ldflags "-X repoDepsCheckerCLI/cmd.Version=$(VERSION)" -o $(BINARY) .

clean:
	rm -f $(BINARY)
