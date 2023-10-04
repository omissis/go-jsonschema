#!/bin/sh -x

set -e
set -o errexit -o nounset

golangci-lint -v run --color=always --config=.rules/.golangci.yml ./...
golangci-lint -v run --color=always --config=.rules/.golangci.yml tests/*.go
golangci-lint -v run --color=always --config=.rules/.golangci.yml tests/helpers/*.go
