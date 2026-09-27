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

_PPROF="http://localhost:6060/debug/pprof"
_BINARY="./bin/oasis-profile"

# CPU profiles and execution traces run back to back, so together
# they cover the whole session
_CPU_SECONDS="30"
_TRACE_SECONDS="30"

# How often memory and goroutine snapshots and process stats are taken
_SNAPSHOT_SECONDS="60"
_PROCESS_SECONDS="5"



##----------------------------------------------------------------------------##
## Arguments                                                                  ##
##----------------------------------------------------------------------------##

while [[ $# -gt 0 ]]; do

	case $1 in

		-h|--help|help)
			cat <<- 'EOF'
			This script builds Oasis and runs it with profiling enabled.
			Profiles are collected in the background for the whole
			session, until Oasis exits or Ctrl+C is pressed. Arguments
			after -- are passed to Oasis.

			Example: ./profile.sh -- --viewer3d

			Everything is written to ./bin/profiles/<date>-<time>:
			  session.txt   Commit, arguments and times of the session
			  oasis.log     Oasis log with frame stats and GC traces
			  cpu/          CPU profiles, 30 seconds each
			  trace/        Execution traces, 30 seconds each
			  heap/         Heap and allocation profiles, every minute
			  goroutine/    Goroutine dumps, every minute
			  process.tsv   CPU and memory use, every 5 seconds

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

for _tool in go curl; do
	if ! command -v "${_tool}" > /dev/null; then
		printf -- "\e[1;31mRequired tool %s is not installed\e[0m\n" "${_tool}" >&2
		exit "${_ERROR_FAILURE}"
	fi
done

# Another process on the pprof port would be profiled instead
if curl -sf -o /dev/null "${_PPROF}/"; then
	printf -- "\e[1;31mPort 6060 is already in use, is Oasis running?\e[0m\n" >&2
	exit "${_ERROR_FAILURE}"
fi



##----------------------------------------------------------------------------##
## Helpers                                                                    ##
##----------------------------------------------------------------------------##

# Saves a capture under a temporary name first, so a capture cut
# off when Oasis exits is never mistaken for a complete one. A
# failed capture waits a second so that retries do not spin.
capture() {

	local url="$1"
	local file="$2"

	if curl -sf -o "${file}.part" "${url}"; then
		mv "${file}.part" "${file}"
	else
		rm -f "${file}.part"
		sleep 1
	fi
}

# Runs back to back CPU profiles while Oasis is running
collect_cpu() {

	local name

	while kill -0 "${_oasis}" 2> /dev/null; do
		name="$(date +%H%M%S)"
		capture "${_PPROF}/profile?seconds=${_CPU_SECONDS}" "${_session}/cpu/${name}.pprof"
	done
}

# Runs back to back execution traces while Oasis is running
collect_trace() {

	local name

	while kill -0 "${_oasis}" 2> /dev/null; do
		name="$(date +%H%M%S)"
		capture "${_PPROF}/trace?seconds=${_TRACE_SECONDS}" "${_session}/trace/${name}.out"
	done
}

# Takes heap, allocation and goroutine snapshots while Oasis is
# running. Allocations are totals since Oasis started.
collect_snapshots() {

	local name

	while kill -0 "${_oasis}" 2> /dev/null; do
		name="$(date +%H%M%S)"

		capture "${_PPROF}/heap" "${_session}/heap/${name}.heap.pprof"
		capture "${_PPROF}/allocs" "${_session}/heap/${name}.allocs.pprof"
		capture "${_PPROF}/goroutine?debug=2" "${_session}/goroutine/${name}.txt"

		sleep "${_SNAPSHOT_SECONDS}"
	done
}

# Records the CPU and memory use of Oasis while it is running
collect_process() {

	local name
	local usage

	printf -- "time\tcpu_percent\trss_kb\n" > "${_session}/process.tsv"

	while kill -0 "${_oasis}" 2> /dev/null; do
		name="$(date +%H:%M:%S)"
		usage="$(ps -o %cpu=,rss= -p "${_oasis}" || true)"

		if [[ -n "${usage}" ]]; then
			# shellcheck disable=SC2086 # Splits the two ps columns
			printf -- "%s\t%s\t%s\n" "${name}" ${usage} >> "${_session}/process.tsv"
		fi

		sleep "${_PROCESS_SECONDS}"
	done
}



##----------------------------------------------------------------------------##
## Main                                                                       ##
##----------------------------------------------------------------------------##

# Change CWD to positions directory
pushd "$(dirname "$0")" > /dev/null

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mBuilding Oasis\e[0m\n"
CGO_ENABLED=0 go build -mod=vendor -tags "viewer3d nofakecgo" -o "${_BINARY}" .

_started="$(date +%Y%m%d-%H%M%S)"
_session="${PWD}/bin/profiles/${_started}"
mkdir -p "${_session}/cpu" "${_session}/trace" "${_session}/heap" "${_session}/goroutine"

_commit="$(git rev-parse --short HEAD 2> /dev/null || printf -- "unknown")"
_changes="$(git status --porcelain 2> /dev/null | wc -l | tr -d ' ')"
_start_time="$(date)"

{
	printf -- "started:   %s\n" "${_start_time}"
	printf -- "commit:    %s (%s files changed)\n" "${_commit}" "${_changes}"
	printf -- "arguments: %s\n" "$*"
} > "${_session}/session.txt"

##----------------------------------------------------------------------------##

printf -- "\n\e[1;32mStarting Oasis\e[0m\n"

# Oasis runs from ./bin, where it finds its offsets and maps.
# GC tracing is enabled for Oasis alone, not the Go toolchain.
pushd ./bin > /dev/null
env GODEBUG=gctrace=1 "../${_BINARY}" --pprof --debug --json "$@" > "${_session}/oasis.log" 2>&1 &
_oasis="$!"
popd > /dev/null

# Wait for the pprof server to come up
_ready="false"

for _ in $(seq 1 60); do
	if curl -sf -o /dev/null "${_PPROF}/"; then
		_ready="true"
		break
	fi

	if ! kill -0 "${_oasis}" 2> /dev/null; then
		break
	fi

	sleep 1
done

if [[ "${_ready}" != "true" ]]; then
	printf -- "\e[1;31mThe pprof server did not start, see %s\e[0m\n" "${_session}/oasis.log" >&2
	kill "${_oasis}" 2> /dev/null || true
	exit "${_ERROR_FAILURE}"
fi

##----------------------------------------------------------------------------##

collect_cpu &
_collectors=("$!")

collect_trace &
_collectors+=("$!")

collect_snapshots &
_collectors+=("$!")

collect_process &
_collectors+=("$!")

printf -- "Profiling into %s\n" "${_session}"
printf -- "Play normally, then quit Oasis or press Ctrl+C to finish\n"

# Ctrl+C also reaches Oasis, which shuts down on its own. The
# script keeps running so it can stop the collectors afterwards.
trap 'printf -- "\nStopping, waiting for Oasis to exit\n"' INT

while kill -0 "${_oasis}" 2> /dev/null; do
	wait "${_oasis}" || true
done

##----------------------------------------------------------------------------##

kill "${_collectors[@]}" 2> /dev/null || true
wait "${_collectors[@]}" 2> /dev/null || true

find "${_session}" -name "*.part" -delete

_end_time="$(date)"
printf -- "ended:     %s\n" "${_end_time}" >> "${_session}/session.txt"

_cpu_count="$(find "${_session}/cpu" -type f | wc -l | tr -d ' ')"
_trace_count="$(find "${_session}/trace" -type f | wc -l | tr -d ' ')"
_size="$(du -sh "${_session}" | cut -f 1)"

printf -- "\n\e[1;32mSession saved to %s\e[0m\n" "${_session}"
printf -- "%s CPU profiles, %s traces, %s in total\n" "${_cpu_count}" "${_trace_count}" "${_size}"

##----------------------------------------------------------------------------##

# Restore the CWD
popd > /dev/null

##----------------------------------------------------------------------------##

} # Ensures the entire script is downloaded
