#!/bin/sh

set -e
set -o errexit -o nounset

CHECKMAKE_VERSION="0.2.2"
GCI_VERSION="v0.13.4"
GINKGO_VERSION="v2.17.1"
GO_COVER_TREEMAP_VERSION="v1.4.2"
GOFUMPT_VERSION="v0.6.0"
GOIMPORTS_VERSION="v0.20.0"
GOLANGCI_LINT_VERSION="v2.10.1"

go install "github.com/daixiang0/gci@${GCI_VERSION}"
go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
go install "github.com/mrtazz/checkmake/cmd/checkmake@${CHECKMAKE_VERSION}"
go install "github.com/nikolaydubina/go-cover-treemap@${GO_COVER_TREEMAP_VERSION}"
go install "github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}"
go install "golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}"
go install "mvdan.cc/gofumpt@${GOFUMPT_VERSION}"
