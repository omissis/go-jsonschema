#!/bin/sh -x

set -e
set -o errexit -o nounset

export GO_VERSION=$(go version | cut -d ' ' -f 3)

goreleaser check
goreleaser --debug release --rm-dist

unset GO_VERSION
