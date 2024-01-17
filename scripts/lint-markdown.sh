#!/bin/sh -x

set -e
set -o errexit -o nounset

markdownlint-cli2 "**/*.md" --config ".rules/.markdownlint.yaml"
