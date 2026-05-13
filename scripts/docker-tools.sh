#!/bin/sh

set -e
set -o errexit -o nounset

docker build --tag tuotuoxp/go-jsonschema:tools-latest --file Dockerfile --target tools .
