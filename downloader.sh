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

_OUTPUT="./bin/depotdownloader"



##----------------------------------------------------------------------------##
## Arguments                                                                  ##
##----------------------------------------------------------------------------##

while [[ $# -gt 0 ]]; do

	case $1 in

		-h|--help|help)
			cat <<- 'EOF'
			This script downloads the latest release of DepotDownloader
			into ./bin/depotdownloader, replacing any older version.
			extract.sh uses it to download the CS2 map files.

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
## Check                                                                      ##
##----------------------------------------------------------------------------##

for _tool in curl unzip; do
	if ! command -v "${_tool}" > /dev/null; then
		printf -- "\e[1;31mRequired tool %s is not installed\e[0m\n" "${_tool}" >&2
		exit "${_ERROR_FAILURE}"
	fi
done

# Release builds are named after the platform they run on
_system="$(uname -s)"
_machine="$(uname -m)"

case "${_system}-${_machine}" in

	Darwin-arm64)  _platform="macos-arm64" ;;
	Darwin-x86_64) _platform="macos-x64"   ;;
	Linux-aarch64) _platform="linux-arm64" ;;
	Linux-x86_64)  _platform="linux-x64"   ;;

	*)
		printf -- "\e[1;31mPlatform %s-%s is not supported\e[0m\n" "${_system}" "${_machine}" >&2
		exit "${_ERROR_FAILURE}"
		;;

esac



##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Change CWD to positions directory
pushd "$(dirname "$0")" > /dev/null

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mDownloading the latest DepotDownloader for %s\e[0m\n" "${_platform}"

rm -rf "${_OUTPUT}"
mkdir -p "${_OUTPUT}"

curl -fsSL -o "${_OUTPUT}/DepotDownloader.zip" \
	"https://github.com/SteamRE/DepotDownloader/releases/latest/download/DepotDownloader-${_platform}.zip"

unzip -q -o "${_OUTPUT}/DepotDownloader.zip" -d "${_OUTPUT}"
rm "${_OUTPUT}/DepotDownloader.zip"

chmod 755 "${_OUTPUT}/DepotDownloader"

# Running it without arguments prints its version first
_version="$("${_OUTPUT}/DepotDownloader" 2>&1 | head -n 1 || true)"
printf -- "Installed %s\n" "${_version}"

##----------------------------------------------------------------------------##

# Restore the CWD
popd > /dev/null

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
