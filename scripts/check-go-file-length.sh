#!/usr/bin/env bash

set -euo pipefail

readonly max_lines=100

mapfile -t files < <(
	find . -type f -name '*.go' \
		! -name '*_test.go' \
		! -path '*/vendor/*' \
		| sort
)

violations=0

for file in "${files[@]}"; do
	line_count="$(wc -l <"$file")"
	line_count="${line_count//[[:space:]]/}"

	if (( line_count > max_lines )); then
		if (( violations == 0 )); then
			echo "Go files longer than ${max_lines} lines:"
		fi

		printf '%s %s\n' "$line_count" "$file"
		violations=$((violations + 1))
	fi
done

if (( violations > 0 )); then
	echo "Found ${violations} file(s) exceeding ${max_lines} lines."
	exit 1
fi

echo "OK: all Go files are ${max_lines} lines or fewer."
