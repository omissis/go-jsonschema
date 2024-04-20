#!/bin/sh

set -e
set -o errexit -o nounset

checkmake --config .rules/checkmake.ini Makefile
