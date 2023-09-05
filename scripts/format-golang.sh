#!/bin/sh -x

set -e
set -o errexit -o nounset

WORKING_DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)

GOFILES=$(find . -name "*.go" -type f -not -path './vendor/*' -not -path './tests/data/*' | sed 's/^\.\///g')

echo "formatting with gofmt.."
echo "${GOFILES}" | xargs -I {} sh -c 'gofmt -w -s {}'

echo "formatting with gofumpt.."
echo "${GOFILES}" | xargs -I {} sh -c 'gofumpt -w -extra {}'

echo "formatting with goimports.."
echo "${GOFILES}" | xargs -I {} sh -c 'goimports -v -w -e -local github.com/atombender {}'

echo "formatting with gci.."
echo "${GOFILES}" | \
xargs -I {} sh -c 'gci write --skip-generated -s standard -s default -s "Prefix(github.com/atombender)" {}'
