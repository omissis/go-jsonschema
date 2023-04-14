#!/bin/sh -x

set -e
set -o errexit -o nounset

yamllint -c .rules/yamllint.yaml .
