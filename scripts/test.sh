#!/bin/sh -x

set -e
set -o errexit -o nounset

go test -v -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
