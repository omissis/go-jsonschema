#!/bin/sh

set -e
set -o errexit -o nounset

docker build --tag tuotuoxp/go-jsonschema:latest --file Dockerfile --target go-jsonschema .
