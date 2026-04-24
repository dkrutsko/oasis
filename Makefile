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

.PHONY: _publish_darwin_amd64 _publish_darwin_arm64 _publish_darwin_universal _publish_linux_amd64 _publish_linux_arm64 _publish_windows_amd64 publish

_publish_darwin_amd64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_darwin_amd64/$(BINARY)"

_publish_darwin_arm64:
	env GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_darwin_arm64/$(BINARY)"

_publish_darwin_universal:
	# Create universal binary using the lipo tool
	mkdir -p "$(OUTPUT)$(BINARY)_darwin_universal"
	lipo -create -output "$(OUTPUT)$(BINARY)_darwin_universal/$(BINARY)" \
		"$(OUTPUT)$(BINARY)_darwin_amd64/$(BINARY)" \
		"$(OUTPUT)$(BINARY)_darwin_arm64/$(BINARY)"
	tar -czf "$(OUTPUT)$(BINARY)_darwin_universal.tar.gz" -C "$(OUTPUT)" "$(BINARY)_darwin_universal"

_publish_linux_amd64:
	env GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_linux_amd64/$(BINARY)"
	tar -czf "$(OUTPUT)$(BINARY)_linux_amd64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_linux_amd64"

_publish_linux_arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_linux_arm64/$(BINARY)"
	tar -czf "$(OUTPUT)$(BINARY)_linux_arm64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_linux_arm64"

_publish_windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_windows_amd64/$(BINARY).exe"
	cd "$(OUTPUT)" && zip -r "$(BINARY)_windows_amd64.zip" "$(BINARY)_windows_amd64"

publish: clean
	$(MAKE) _publish_darwin_amd64 &
	$(MAKE) _publish_darwin_arm64 &
	$(MAKE) _publish_linux_amd64 &
	$(MAKE) _publish_linux_arm64 &
	$(MAKE) _publish_windows_amd64 &
	wait
	$(MAKE) _publish_darwin_universal
