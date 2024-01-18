#!/bin/sh -x

set -e
set -o errexit -o nounset

find . -name "*.json" -type f -exec sh -c 'jq -M . "$1" > "$1".tmp' shell {} \; -exec mv {}.tmp {} \;
