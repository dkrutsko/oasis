#!/bin/bash
# shellcheck enable=all

{ # Ensures the entire script is downloaded

# Use strict mode
set -euo pipefail

##----------------------------------------------------------------------------##
## Constants                                                                  ##
##----------------------------------------------------------------------------##

_ERROR_SUCCESS="0"
_ERROR_FAILURE="1"
_ERROR_OPTIONS="2"



##----------------------------------------------------------------------------##
## Arguments                                                                  ##
##----------------------------------------------------------------------------##

while [[ $# -gt 0 ]]; do

	case $1 in

		-h|--help|help)
			cat <<- 'EOF'
			This script downloads and saves the latest offsets from cs2-dumper.

			EOF
			exit "${_ERROR_SUCCESS}"
			;;

		# End of args
		--)
			shift
			break
			;;

		# Unknown arg
		*)
			printf -- "\e[0;31mArgument \e[1;31m%s\e[0;31m is invalid\e[0m\n" "$1" >&2
			exit "${_ERROR_OPTIONS}"
			;;

	esac

done



##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Change CWD to positions directory
pushd "$(dirname "$0")" > /dev/null

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mEnsuring output directory\e[0m\n"
mkdir -p "./static"

printf -- "\n\e[1;32mDownloading offsets.json\e[0m\n"
curl -sSL -o "./static/offsets.json" "https://raw.githubusercontent.com/a2x/cs2-dumper/main/output/offsets.json"

##----------------------------------------------------------------------------##

# Restore the CWD
popd > /dev/null

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
