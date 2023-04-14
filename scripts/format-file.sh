#!/bin/sh -x

set -e
set -o errexit -o nounset

IGNORE_PATHS='.git/,.github/,.vscode/,.idea/'

file-cr --text --ignore "${IGNORE_PATHS}" --fix --path .
file-crlf --text --ignore "${IGNORE_PATHS}" --fix --path .
file-nullbyte --text --ignore "${IGNORE_PATHS}" --fix --path .
file-trailing-newline --text --ignore "${IGNORE_PATHS}" --fix --path .
file-trailing-single-newline --text --ignore "${IGNORE_PATHS}" --fix --path .
file-trailing-space --text --ignore "${IGNORE_PATHS}" --fix --path .
file-utf8 --text --ignore "${IGNORE_PATHS}" --fix --path .
file-utf8-bom --text --ignore "${IGNORE_PATHS}" --fix --path .
