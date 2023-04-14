#!/bin/sh -x

set -e
set -o errexit -o nounset

export _BIN_OPEN="open"
if [ "$(uname -s)" = "Linux" ]; then
    _BIN_OPEN="xdg-open"
fi

go tool cover -html=coverage.out -o coverage.html

go-cover-treemap -coverprofile coverage.out > coverage.svg && ${_BIN_OPEN} coverage.svg

unset _BIN_OPEN
