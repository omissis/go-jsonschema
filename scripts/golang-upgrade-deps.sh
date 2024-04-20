#!/bin/sh

set -e
set -o errexit -o nounset

go get -u ./... && go mod tidy
