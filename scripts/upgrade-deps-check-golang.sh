#!/bin/sh -x

set -e
set -o errexit -o nounset

go list -mod=readonly -u \
    -f "{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}" \
    -m all
