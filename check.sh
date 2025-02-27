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

# Run through all the unit tests in the project
printf -- "\n\e[1;32mRunning unit tests\e[0m\n"
make test

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
