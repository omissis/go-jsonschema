#!/bin/sh

set -e
set -o errexit -o nounset

shfmt -i 2 -ci -sr -w .
