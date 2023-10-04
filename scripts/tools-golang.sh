#!/bin/sh -x

set -e
set -o errexit -o nounset

CHECKMAKE_VERSION="0.2.1"
GCI_VERSION="v0.10.1"
GINKGO_VERSION="v2.9.2"
GO_COVER_TREEMAP_VERSION="v1.3.0"
GOFUMPT_VERSION="v0.5.0"
GOIMPORTS_VERSION="v0.8.0"
GOLANGCI_LINT_VERSION="v1.54.2"

go install "github.com/daixiang0/gci@${GCI_VERSION}"
go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
go install "github.com/mrtazz/checkmake/cmd/checkmake@${CHECKMAKE_VERSION}"
go install "github.com/nikolaydubina/go-cover-treemap@${GO_COVER_TREEMAP_VERSION}"
go install "github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}"
go install "golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}"
go install "mvdan.cc/gofumpt@${GOFUMPT_VERSION}"
