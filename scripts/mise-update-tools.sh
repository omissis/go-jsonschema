#!/bin/sh

# shellcheck disable=SC2016

set -e
set -o errexit -o nounset

yq '.tools | keys | .[]' mise.toml -oy |
  xargs -I {} -n 1 sh -c 'echo {} $(mise latest {} 2>/dev/null)' > \
    mise.toml.tmp &&
  mv mise.toml.tmp mise.toml
