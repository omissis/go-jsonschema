#!/bin/sh

set -e
set -o errexit -o nounset

mkdir -p "${PWD}/coverage/pkg"
mkdir -p "${PWD}/coverage/tests"

go test -v -race -mod=readonly -covermode=atomic -coverpkg=./... -cover ./... -args -test.gocoverdir="${PWD}/coverage/pkg"
go test -v -race -mod=readonly -covermode=atomic -coverpkg=./... -cover ./tests -args -test.gocoverdir="${PWD}/coverage/tests"

go tool covdata textfmt -i=./coverage/tests,./coverage/pkg -o coverage.out
