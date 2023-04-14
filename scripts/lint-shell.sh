#!/bin/sh -x

set -e
set -o errexit -o nounset

shellcheck -a -o all -s bash -- **/*.sh
