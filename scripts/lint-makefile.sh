#!/bin/sh -x

set -e
set -o errexit -o nounset

checkmake --config .rules/checkmake.ini Makefile
