#!/bin/sh
set -eu

raw_profile="$(mktemp)"
filtered_profile="$(mktemp)"
trap 'rm -f "$raw_profile" "$filtered_profile"' EXIT

go test -count=1 ./... -coverprofile="$raw_profile"

awk '
NR == 1 { print; next }
$1 !~ /\/docs\/swagger\/docs.go:/ && $1 !~ /\/cmd\/api\/main.go:/ { print }
' "$raw_profile" > "$filtered_profile"

go tool cover -func="$filtered_profile"

if awk 'NR > 1 && $2 > 0 && $3 == 0 { uncovered = 1 } END { exit !uncovered }' "$filtered_profile"; then
	echo "uncovered useful-code blocks:" >&2
	awk 'NR > 1 && $2 > 0 && $3 == 0 { print }' "$filtered_profile" >&2
	exit 1
fi

coverage="$(go tool cover -func="$filtered_profile" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"

if [ "$coverage" != "100.0" ]; then
	echo "useful-code coverage is ${coverage}%; expected 100.0%" >&2
	exit 1
fi
