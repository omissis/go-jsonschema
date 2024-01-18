#!/bin/sh -x

set -e
set -o errexit -o nounset

GO_VERSION=$(go version | cut -d ' ' -f 3)

export GO_VERSION

goreleaser check
goreleaser release --debug --snapshot --clean

unset GO_VERSION
