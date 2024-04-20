#!/bin/sh

set -e
set -o errexit -o nounset

yamllint -c .rules/yamllint.yaml .
