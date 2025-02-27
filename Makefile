##----------------------------------------------------------------------------##
## Variables                                                                  ##
##----------------------------------------------------------------------------##

OUTPUT = ./bin/
BINARY = oasis



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
	@echo "  $$ make test    - Runs various unit tests on project"
	@echo "  $$ make clean   - Cleans and removes generated files"
	@echo "  $$ make publish - Builds artifacts for a new release"
	@echo
	@echo "DOCS"
	@echo "  Visit https://github.com/dkrutsko/oasis for more"
	@echo



##----------------------------------------------------------------------------##
## Build                                                                      ##
##----------------------------------------------------------------------------##

.PHONY: build debug test clean

build:
	go build -ldflags "-s -w" -o "$(OUTPUT)$(BINARY)"

debug:
	# Include -gcflags to improve the experience with GDB
	# particularly when needing to print variable values.
	# Also try to detect race conditions using -race flag.
	go build -o "$(OUTPUT)$(BINARY)" -gcflags="all=-N -l" -race

test:
	# Do test without caching
	go test -count=1 ./utility -race

clean:
	rm -rf "$(OUTPUT)"



##----------------------------------------------------------------------------##
## Publish                                                                    ##
##----------------------------------------------------------------------------##

.PHONY: publish

publish: clean
	env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o "$(OUTPUT)$(BINARY).exe"
