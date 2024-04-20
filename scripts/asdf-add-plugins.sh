#!/bin/sh

set -e
set -o errexit -o nounset

cut -d ' ' -f 1 < .tool-versions | xargs -n 1 asdf plugin add
