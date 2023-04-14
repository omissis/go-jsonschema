#!/bin/sh -x

set -e
set -o errexit -o nounset

find . \
  -type f \
  -not -path ".git" \
  -not -path ".github" \
  -not -path ".vscode" \
  -not -path ".idea" \
  -name "*.json" \
  -exec jsonlint -c -q -t '  ' {} \;
