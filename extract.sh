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

# Steam app and depot that hold the CS2 map files. The depot
# downloads without a Steam login.
_APP="730"
_DEPOT="2347770"

# Extractor exit code for VPKs without world physics, such as
# the vanity and settings packs in the maps folder
_EXTRACTOR_NO_PHYSICS="3"

_OUTPUT="./bin/maps"
_MANIFEST="${_OUTPUT}/manifest.json"
_DOWNLOADS="${_OUTPUT}/downloads"
_EXTRACTOR="./bin/extractor"
_DOWNLOADER="./bin/depotdownloader/DepotDownloader"



##----------------------------------------------------------------------------##
## Arguments                                                                  ##
##----------------------------------------------------------------------------##

_force="false"

while [[ $# -gt 0 ]]; do

	case $1 in

		-h|--help|help)
			cat <<- 'EOF'
			This script downloads the CS2 maps that changed since the
			last run, one at a time, extracts their collision geometry
			and compresses it into ./bin/maps for Oasis to load. Run
			./downloader.sh first to download DepotDownloader.

			The manifest in ./bin/maps records the checksum of each
			map's VPK. A map is extracted again when its VPK changes
			or when its .tri.zst file is deleted.

			--force extracts every map again, even unchanged ones.

			EOF
			exit "${_ERROR_SUCCESS}"
			;;

		--force)
			_force="true"
			shift
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

for _tool in make go zstd jq shasum; do
	if ! command -v "${_tool}" > /dev/null; then
		printf -- "\e[1;31mRequired tool %s is not installed\e[0m\n" "${_tool}" >&2
		exit "${_ERROR_FAILURE}"
	fi
done



##----------------------------------------------------------------------------##
## Helpers                                                                    ##
##----------------------------------------------------------------------------##

# Prints the VPK checksum and result the manifest holds for a
# map, or nothing when the map has no entry yet
get_manifest_map() {

	jq -r --arg name "$1" '.maps[$name] // empty | "\(.sha1) \(.result)"' "${_MANIFEST}"
}

# Records the VPK checksum and result of a map. The file is
# replaced in one step so that an interrupted run never leaves
# it half written.
set_manifest_map() {

	jq -S --arg name "$1" --arg sha1 "$2" --arg result "$3" \
		'.maps[$name] = {sha1: $sha1, result: $result}' "${_MANIFEST}" > "${_MANIFEST}.tmp"

	mv "${_MANIFEST}.tmp" "${_MANIFEST}"
}



##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Change CWD to positions directory
pushd "$(dirname "$0")" > /dev/null

##----------------------------------------------------------------------------##

if [[ ! -f "${_DOWNLOADER}" ]]; then
	printf -- "\e[1;31mDepotDownloader is missing, run ./downloader.sh to download it\e[0m\n" >&2
	exit "${_ERROR_FAILURE}"
fi

# check.sh resets the permissions of every file in the repository
chmod 755 "${_DOWNLOADER}"

printf -- "\n\e[1;32mBuilding the extractor\e[0m\n"
make extractor > /dev/null

##----------------------------------------------------------------------------##

mkdir -p "${_OUTPUT}"

if [[ -f "${_MANIFEST}" ]]; then
	printf -- "\n\e[1;32mUsing the manifest from the last run\e[0m\n"
else
	printf -- "\n\e[1;32mNo manifest yet, extracting every map\e[0m\n"
	jq -n '{maps: {}}' > "${_MANIFEST}"
fi

# The file list of the latest depot manifest has the SHA-1 of
# every file, so changed maps are found without downloading them
printf -- "\n\e[1;32mDownloading the depot file list\e[0m\n"

rm -rf "${_DOWNLOADS}"
mkdir -p "${_DOWNLOADS}"

if ! "${_DOWNLOADER}" -app "${_APP}" -depot "${_DEPOT}" -manifest-only -dir "${_DOWNLOADS}" \
	> "${_DOWNLOADS}/downloader.log" 2>&1 < /dev/null; then
	printf -- "\e[1;31mFailed to download the depot file list\e[0m\n" >&2
	tail -n 20 "${_DOWNLOADS}/downloader.log" >&2
	exit "${_ERROR_FAILURE}"
fi

_file_list=""
for _file in "${_DOWNLOADS}/manifest_${_DEPOT}_"*.txt; do
	_file_list="${_file}"
done

if [[ ! -f "${_file_list}" ]]; then
	printf -- "\e[1;31mDepot file list is missing\e[0m\n" >&2
	exit "${_ERROR_FAILURE}"
fi

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mFinding the maps to extract\e[0m\n"

# Each map line holds the SHA-1 and path of a map VPK
_maps="${_DOWNLOADS}/maps.txt"
awk '$NF ~ /^game\/csgo\/maps\/[^\/]+\.vpk$/ { print $3, $NF }' "${_file_list}" > "${_maps}"

# Each work line holds a map name and the SHA-1 of its VPK
_work="${_DOWNLOADS}/work.txt"
: > "${_work}"

while read -r _sha1 _path <&3; do

	_name="$(basename "${_path}" .vpk)"
	_entry="$(get_manifest_map "${_name}")"

	_last_sha1=""
	_last_result=""
	read -r _last_sha1 _last_result <<< "${_entry}" || true

	_needed="false"

	if [[ "${_force}" == "true" || "${_sha1}" != "${_last_sha1}" ]]; then
		_needed="true"
	elif [[ "${_last_result}" == "extracted" && ! -f "${_OUTPUT}/${_name}.tri.zst" ]]; then
		_needed="true"
	fi

	if [[ "${_needed}" == "true" ]]; then
		printf -- "%s %s\n" "${_name}" "${_sha1}" >> "${_work}"
	fi

done 3< "${_maps}"

_total="$(wc -l < "${_work}" | tr -d ' ')"
printf -- "%s maps to extract\n" "${_total}"

##----------------------------------------------------------------------------##

_index="0"
_failed="0"

while read -r _name _sha1 <&3; do

	_index="$(( _index + 1 ))"
	printf -- "\n\e[1;32m[%s/%s] %s\e[0m\n" "${_index}" "${_total}" "${_name}"

	# Download the VPK on its own
	_vpk="${_DOWNLOADS}/game/csgo/maps/${_name}.vpk"
	printf -- "game/csgo/maps/%s.vpk\n" "${_name}" > "${_DOWNLOADS}/filelist.txt"

	if ! "${_DOWNLOADER}" -app "${_APP}" -depot "${_DEPOT}" -filelist "${_DOWNLOADS}/filelist.txt" \
		-dir "${_DOWNLOADS}" > "${_DOWNLOADS}/downloader.log" 2>&1 < /dev/null || [[ ! -f "${_vpk}" ]]; then
		printf -- "\e[1;31mFailed to download %s\e[0m\n" "${_name}" >&2
		tail -n 20 "${_DOWNLOADS}/downloader.log" >&2
		_failed="$(( _failed + 1 ))"
		continue
	fi

	# The depot can update between the file list and the download
	_actual_sha1="$(shasum -a 1 "${_vpk}" | cut -d ' ' -f 1)"

	if [[ "${_actual_sha1}" != "${_sha1}" ]]; then
		printf -- "\e[1;31mChecksum of %s does not match the file list\e[0m\n" "${_name}" >&2
		rm -f "${_vpk}"
		_failed="$(( _failed + 1 ))"
		continue
	fi

	# Extract the collision geometry, then drop the VPK
	_code="0"
	"${_EXTRACTOR}" "${_vpk}" "${_OUTPUT}" > "${_DOWNLOADS}/extractor.log" 2>&1 || _code="$?"
	rm -f "${_vpk}"

	if [[ "${_code}" == "${_EXTRACTOR_NO_PHYSICS}" ]]; then
		printf -- "No world physics, skipped\n"
		set_manifest_map "${_name}" "${_sha1}" "no_physics"
		continue
	fi

	if [[ "${_code}" != "0" ]]; then
		printf -- "\e[1;31mFailed to extract %s\e[0m\n" "${_name}" >&2
		cat "${_DOWNLOADS}/extractor.log" >&2
		_failed="$(( _failed + 1 ))"
		continue
	fi

	# Compress each extracted map, which removes its .tri file
	for _tri in "${_OUTPUT}"/*.tri; do
		[[ -f "${_tri}" ]] || continue

		zstd -19 -T4 -q -f --rm "${_tri}" -o "${_tri}.zst"

		_size="$(wc -c < "${_tri}.zst" | tr -d ' ')"
		_size="$(awk -v size="${_size}" 'BEGIN { printf "%.1f", size / 1000000 }')"
		printf -- "Wrote %s (%s MB)\n" "${_tri}.zst" "${_size}"
	done

	set_manifest_map "${_name}" "${_sha1}" "extracted"

done 3< "${_work}"

rm -rf "${_DOWNLOADS}"

##----------------------------------------------------------------------------##

# Failed maps keep their old manifest entry, so they are tried
# again on the next run
if [[ "${_failed}" != "0" ]]; then
	printf -- "\n\e[1;31m%s maps failed and will be retried on the next run\e[0m\n" "${_failed}" >&2
	exit "${_ERROR_FAILURE}"
fi

printf -- "\n\e[1;32mThe maps are up to date\e[0m\n"

##----------------------------------------------------------------------------##

# Restore the CWD
popd > /dev/null

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
