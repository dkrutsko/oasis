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
			This script compiles a production version of this software and
			copies the binary and its runtime configuration to a directory.
			--dir specifies the directory to copy the data to.

			EOF
			exit "${_ERROR_SUCCESS}"
			;;

		--dir)
			_dir="$2"
			shift # arg
			shift # val
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

# Ensure directory specified
if [[ -z "${_dir:-}" ]]; then

	printf -- "\e[1;31mSpecify valid directory to write to\e[0m\n" >&2
	exit "${_ERROR_FAILURE}"

fi



##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Change CWD to positions directory
pushd "$(dirname "$0")" > /dev/null

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mEnsuring output directory\e[0m\n"
mkdir -p "${_dir}"

printf -- "\n\e[1;32mBuilding Windows binary\e[0m\n"
make publish

printf -- "\n\e[1;32mCopying runtime to host\e[0m\n"
cp -r ./bin/* "${_dir}" # TODO: Switch this to rsync with --delete when entire runtime is present

##----------------------------------------------------------------------------##

# Restore the CWD
popd > /dev/null

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
