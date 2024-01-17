#!/bin/sh -x

set -e
set -o errexit -o nounset

find . \( -name '*.yaml' -o -name '*.yml' \) -type f -exec yq eval -P -I 2 -M -i {} \;
