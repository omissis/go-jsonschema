#!/bin/sh -x

set -e
set -o errexit -o nounset

find . -name *.json -type f -exec jq -M . {} > {}.tmp \; -exec mv {}.tmp {} \;
