##----------------------------------------------------------------------------##
## Variables                                                                  ##
##----------------------------------------------------------------------------##

OUTPUT = ./bin/
BINARY = oasis
EXTRACTOR = extractor

# Host platform name, used to pick the matching redistributable
# libraries from redist/ for a local build. The macOS libraries
# are universal, so both Apple architectures share "darwin".
HOST_OS   = $(shell go env GOOS)
HOST_ARCH = $(shell go env GOARCH)
ifeq ($(HOST_OS),darwin)
PLATFORM = darwin
else
PLATFORM = $(HOST_OS)_$(HOST_ARCH)
endif

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
	@echo "  $$ make help      - Prints out these help instructions"
	@echo "  $$ make build     - Builds main binary in release mode"
	@echo "  $$ make debug     - Builds main binary in debug mode"
	@echo "  $$ make sync      - Copies platform libraries to the output dir"
	@echo "  $$ make clean     - Cleans and removes generated files"
	@echo "  $$ make publish   - Builds release archives for all platforms"
	@echo "  $$ make extractor - Builds the map extractor binary"
	@echo
	@echo "DOCS"
	@echo "  Visit https://github.com/dkrutsko/oasis for more"
	@echo



##----------------------------------------------------------------------------##
## Build                                                                      ##
##----------------------------------------------------------------------------##

.PHONY: build debug sync clean

build: sync
	go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)"

debug: sync
	# Include -gcflags to improve the experience with GDB
	# particularly when needing to print variable values.
	# Also try to detect race conditions using -race flag.
	go build -ldflags "$(LDFLAGS)" -o "$(OUTPUT)$(BINARY)" -gcflags="all=-N -l" -race

# Copy the current platform's redistributable libraries into the
# output directory so a locally built binary can run in place.
# bin/ is throwaway build output removed by clean, so the redist
# files live under redist/ and are synced here on demand.
sync:
	mkdir -p "$(OUTPUT)"
	rsync -a redist/$(PLATFORM)/ "$(OUTPUT)"

clean:
	rm -rf "$(OUTPUT)"



##----------------------------------------------------------------------------##
## Extractor                                                                  ##
##----------------------------------------------------------------------------##

.PHONY: extractor extractor_linux_amd64

extractor:
	go build -ldflags "-s -w" -o "$(OUTPUT)$(EXTRACTOR)" ./extractor

extractor_linux_amd64:
	# Static build for the map pipeline container
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "$(OUTPUT)$(EXTRACTOR)_linux_amd64/$(EXTRACTOR)" ./extractor



##----------------------------------------------------------------------------##
## Publish                                                                    ##
##----------------------------------------------------------------------------##

.PHONY: _publish_darwin_amd64 _publish_darwin_arm64 _publish_linux_amd64 _publish_linux_arm64 _publish_windows_amd64 publish

# Each target builds the binary for one platform, syncs that
# platform's redistributable libraries and license alongside it,
# then archives the staging directory for a GitHub release. The
# archive layout mirrors a local build: binary next to libraries.

_publish_darwin_amd64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_darwin_amd64/$(BINARY)"
	rsync -a redist/darwin/ "$(OUTPUT)$(BINARY)_darwin_amd64"
	tar -czf "$(OUTPUT)$(BINARY)_darwin_amd64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_darwin_amd64"

_publish_darwin_arm64:
	env GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_darwin_arm64/$(BINARY)"
	rsync -a redist/darwin/ "$(OUTPUT)$(BINARY)_darwin_arm64"
	tar -czf "$(OUTPUT)$(BINARY)_darwin_arm64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_darwin_arm64"

_publish_linux_amd64:
	env GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_linux_amd64/$(BINARY)"
	rsync -a redist/linux_amd64/ "$(OUTPUT)$(BINARY)_linux_amd64"
	tar -czf "$(OUTPUT)$(BINARY)_linux_amd64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_linux_amd64"

_publish_linux_arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_linux_arm64/$(BINARY)"
	rsync -a redist/linux_arm64/ "$(OUTPUT)$(BINARY)_linux_arm64"
	tar -czf "$(OUTPUT)$(BINARY)_linux_arm64.tar.gz" -C "$(OUTPUT)" "$(BINARY)_linux_arm64"

_publish_windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" -o "$(OUTPUT)$(BINARY)_windows_amd64/$(BINARY).exe"
	rsync -a redist/windows_amd64/ "$(OUTPUT)$(BINARY)_windows_amd64"
	cd "$(OUTPUT)" && zip -r "$(BINARY)_windows_amd64.zip" "$(BINARY)_windows_amd64"

publish: clean
	$(MAKE) _publish_darwin_amd64
	$(MAKE) _publish_darwin_arm64
	$(MAKE) _publish_linux_amd64
	$(MAKE) _publish_linux_arm64
	$(MAKE) _publish_windows_amd64
