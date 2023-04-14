#!/bin/sh -x

set -e
set -o errexit -o nounset

find . \
  -type f \
  -name '*Dockerfile*' \
  -not -path './.git/*' \
  -exec hadolint {} \;
