#!/bin/sh
set -eu

minimum_coverage="${COVERAGE_MINIMUM:-40.0}"
coverage_profile="$(mktemp)"
trap 'rm -f "$coverage_profile"' EXIT

go test -count=1 -covermode=atomic -coverpkg=./... -coverprofile="$coverage_profile" ./...
go tool cover -func="$coverage_profile"

coverage="$(go tool cover -func="$coverage_profile" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"

if ! awk -v coverage="$coverage" -v minimum="$minimum_coverage" 'BEGIN { exit !(coverage >= minimum) }'; then
	echo "global statement coverage is ${coverage}%; expected at least ${minimum_coverage}%" >&2
	exit 1
fi

echo "global statement coverage is ${coverage}% (minimum ${minimum_coverage}%)"
