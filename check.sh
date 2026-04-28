#!/bin/bash
# shellcheck enable=all

{ # Ensures the entire script is downloaded

# Use strict mode
set -euo pipefail

##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Ensure that all the files have the correct permissions
printf -- "\n\e[1;32mNormalizing file permissions\e[0m\n"
find . -type f ! -ipath "./.git*/*" -exec chmod 644 {} \;
find . -type f   -iname "*.sh"      -exec chmod 755 {} \;

# Perform linting on every script in the project
printf -- "\n\e[1;32mLinting all scripts\e[0m\n"
find . -type f -iname "*.sh" -exec shellcheck {} +

##----------------------------------------------------------------------------##

# Ensure all the Go module dependencies are tidy
printf -- "\n\e[1;32mTidying Go modules\e[0m\n"
go mod tidy

# Format and simplify all Go code in the project
printf -- "\n\e[1;32mFormatting Go code\e[0m\n"
gofmt -s -w .

# Run static analysis checks on all Go packages. The unsafeptr
# analyzer is disabled because the leech package requires the
# uintptr(unsafe.Pointer(...)) pattern for DLL interop.
printf -- "\n\e[1;32mRunning static analysis\e[0m\n"
go vet -unsafeptr=false ./...

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
