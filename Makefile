##----------------------------------------------------------------------------##
## Variables                                                                  ##
##----------------------------------------------------------------------------##

OUTPUT = ./bin/
BINARY = oasis

LDFLAGS = -X 'github.com/dkrutsko/oasis/config.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedCommit=$(shell git --no-pager rev-parse --verify HEAD | base64)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedDate=$(shell git --no-pager show -s --format=%aI | base64)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedMessage=$(shell git --no-pager log -1 --pretty=%B | base64)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedRemote=$(shell git --no-pager ls-remote --get-url | base64)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedBranch=$(shell git --no-pager branch --show-current | base64)' \
          -X 'github.com/dkrutsko/oasis/utility.gitEmbedStatus=$(shell git --no-pager status --porcelain | base64)'



##----------------------------------------------------------------------------##
## Help                                                                       ##
##----------------------------------------------------------------------------##

.PHONY: help

help:
	@echo
	@echo "WELCOME TO OASIS"
	@echo "----------------"
	@echo
	@echo "MAKE"
	@echo "  $$ make help    - Prints out these help instructions"
	@echo "  $$ make build   - Builds main binary in release mode"
	@echo "  $$ make debug   - Builds main binary in debug mode"
	@echo "  $$ make clean   - Cleans and removes generated files"
	@echo "  $$ make publish - Builds artifacts for a new release"
	@echo
	@echo "DOCS"
	@echo "  Visit https://github.com/dkrutsko/oasis for more"
	@echo



##----------------------------------------------------------------------------##
## Build                                                                      ##
##----------------------------------------------------------------------------##

.PHONY: build debug clean

build:
	go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)"

debug:
	# Include -gcflags to improve the experience with GDB
	# particularly when needing to print variable values.
	# Also try to detect race conditions using -race flag.
	go build -ldflags "$(LDFLAGS)" -o "$(OUTPUT)$(BINARY)" -gcflags="all=-N -l" -race

clean:
	rm -rf "$(OUTPUT)"



##----------------------------------------------------------------------------##
## Publish                                                                    ##
##----------------------------------------------------------------------------##

.PHONY: publish

publish: clean
	env GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY).exe"
