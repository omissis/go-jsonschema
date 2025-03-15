ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SHELL := /bin/bash

.DEFAULT_GOAL := help

# ----------------------------------------------------------------------------------------------------------------------
# Private variables
# ----------------------------------------------------------------------------------------------------------------------

_DOCKER_FILELINT_IMAGE=cytopia/file-lint:latest-0.8
_DOCKER_GOLANG_IMAGE=golang:1.23.7
_DOCKER_GOLANGCI_LINT_IMAGE=golangci/golangci-lint:v1.62.2
_DOCKER_HADOLINT_IMAGE=hadolint/hadolint:v2.12.0
_DOCKER_JSONLINT_IMAGE=cytopia/jsonlint:1.6
_DOCKER_MAKEFILELINT_IMAGE=cytopia/checkmake:latest-0.5
_DOCKER_MARKDOWNLINT_IMAGE=davidanson/markdownlint-cli2:v0.14.0
_DOCKER_SHELLCHECK_IMAGE=koalaman/shellcheck-alpine:v0.10.0
_DOCKER_SHFMT_IMAGE=mvdan/shfmt:v3-alpine
_DOCKER_TOOLS_IMAGE=omissis/go-jsonschema:tools-latest
_DOCKER_YAMLLINT_IMAGE=cytopia/yamllint:1

_PROJECT_DIRECTORY=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

# ----------------------------------------------------------------------------------------------------------------------
# Utility functions
# ----------------------------------------------------------------------------------------------------------------------

#1: docker image
#2: script name
define run-script-docker
	@docker run --rm \
		-v ${_PROJECT_DIRECTORY}:/data \
		-w /data \
		--entrypoint /bin/sh \
		$(1) scripts/$(2).sh
endef

# check-variable-%: Check if the variable is defined.
check-variable-%:
	@[[ "${${*}}" ]] || (echo '*** Please define variable `${*}` ***' && exit 1)

# ----------------------------------------------------------------------------------------------------------------------
# Linting Targets
# ----------------------------------------------------------------------------------------------------------------------

.PHONY: lint lint-docker

lint: lint-markdown lint-shell lint-yaml lint-dockerfile lint-makefile lint-json lint-file lint-go
lint-docker: lint-markdown-docker lint-shell-docker lint-yaml-docker lint-dockerfile-docker lint-makefile-docker lint-json-docker lint-file-docker lint-go-docker

.PHONY: lint-markdown lint-markdown-docker

lint-markdown:
	@scripts/lint-markdown.sh

lint-markdown-docker:
	$(call run-script-docker,${_DOCKER_MARKDOWNLINT_IMAGE},lint-markdown)

.PHONY: lint-shell lint-shell-docker

lint-shell:
	@scripts/lint-shell.sh

lint-shell-docker:
	$(call run-script-docker,${_DOCKER_SHELLCHECK_IMAGE},lint-shell)

.PHONY: lint-yaml lint-yaml-docker

lint-yaml:
	@scripts/lint-yaml.sh

lint-yaml-docker:
	$(call run-script-docker,${_DOCKER_YAMLLINT_IMAGE},lint-yaml)

.PHONY: lint-dockerfile lint-dockerfile-docker

lint-dockerfile:
	@scripts/lint-dockerfile.sh

lint-dockerfile-docker:
	$(call run-script-docker,${_DOCKER_HADOLINT_IMAGE},lint-dockerfile)

.PHONY: lint-makefile lint-makefile-docker

lint-makefile:
	@scripts/lint-makefile.sh

lint-makefile-docker:
	$(call run-script-docker,${_DOCKER_MAKEFILELINT_IMAGE},lint-makefile)

.PHONY: lint-json lint-json-docker

lint-json:
	@scripts/lint-json.sh

lint-json-docker:
	$(call run-script-docker,${_DOCKER_JSONLINT_IMAGE},lint-json)

.PHONY: lint-file lint-file-docker

lint-file:
	@scripts/lint-file.sh

lint-file-docker:
	$(call run-script-docker,${_DOCKER_FILELINT_IMAGE},lint-file)

# ----------------------------------------------------------------------------------------------------------------------
# Formatting Targets
# ----------------------------------------------------------------------------------------------------------------------

.PHONY: format format-docker

format: format-file format-shell format-markdown format-go
format-docker: format-file-docker format-shell-docker format-markdown-docker format-go-docker

.PHONY: format-file format-file-docker

format-file:
	@scripts/format-file.sh

format-file-docker:
	$(call run-script-docker,${_DOCKER_FILELINT_IMAGE},format-file)

.PHONY: format-markdown format-markdown-docker

format-markdown:
	@scripts/format-markdown.sh

format-markdown-docker:
	$(call run-script-docker,${_DOCKER_MARKDOWNLINT_IMAGE},format-markdown)

.PHONY: format-shell format-shell-docker

format-shell:
	@scripts/format-shell.sh

format-shell-docker:
	$(call run-script-docker,${_DOCKER_SHFMT_IMAGE},format-shell)

.PHONY: format-yaml format-yaml-docker

format-yaml:
	@scripts/format-yaml.sh

format-yaml-docker: docker-tools
	$(call run-script-docker,${_DOCKER_TOOLS_IMAGE},format-yaml)

.PHONY: format-json format-json-docker

format-json:
	@scripts/format-json.sh

format-json-docker: docker-tools
	$(call run-script-docker,${_DOCKER_TOOLS_IMAGE},format-json)

# ----------------------------------------------------------------------------------------------------------------------
# Development Targets
# ----------------------------------------------------------------------------------------------------------------------

.PHONY: env

env:
	@echo 'export CGO_ENABLED=0'
	@echo 'export GOARCH=${_GOARCH}'
	@grep -v '^#' .env | sed 's/^/export /'

.PHONY: docker-build docker-tools

docker-build:
	@scripts/docker-build.sh

docker-tools:
	@scripts/docker-tools.sh

.PHONY: asdf-install-tools asdf-update-tools

asdf-install-tools:
	@scripts/asdf-add-plugins.sh
	@asdf install

asdf-update-tools:
	@scripts/asdf-update-tools.sh
	@asdf install

.PHONY: golang-check-updates golang-update golang-install-tools

golang-check-update-deps:
	@scripts/upgrade-deps-check-golang.sh

golang-update-deps:
	@scripts/upgrade-deps-golang.sh

golang-install-tools:
	@scripts/golang-install-tools.sh

# ----------------------------------------------------------------------------------------------------------------------
# Golang Targets
# ----------------------------------------------------------------------------------------------------------------------

.PHONY: tools-go tools-brew

tools-go:
	@scripts/golang-install-tools.sh

tools-brew:
	@scripts/tools-brew.sh

.PHONY: lint-go lint-go-docker

lint-go:
	@scripts/lint-golang.sh

lint-go-docker:
	$(call run-script-docker,${_DOCKER_GOLANGCI_LINT_IMAGE},lint-golang)

.PHONY: format-go format-go-docker

format-go:
	@scripts/format-golang.sh

format-go-docker: docker-tools
	$(call run-script-docker,${_DOCKER_TOOLS_IMAGE},format-golang)

.PHONY: upgrade-deps-check-go upgrade-deps-check-go-docker upgrade-deps-go upgrade-deps-go-docker

upgrade-deps-check-go:
	@scripts/upgrade-deps-check-golang.sh

upgrade-deps-check-go-docker:
	$(call run-script-docker,${_DOCKER_GOLANG_IMAGE},upgrade-deps-check-golang)

upgrade-deps-go:
	@scripts/upgrade-deps-golang.sh

upgrade-deps-go-docker:
	$(call run-script-docker,${_DOCKER_GOLANG_IMAGE},upgrade-deps-golang)

.PHONY: test test-docker show-coverage-go

test:
	@scripts/test.sh

test-docker: docker-tools
	$(call run-script-docker,${_DOCKER_TOOLS_IMAGE},test)

show-coverage-go:
	@scripts/show-coverage-golang.sh

.PHONY: build release

build:
	@scripts/build.sh

release:
	@scripts/release.sh

# ----------------------------------------------------------------------------------------------------------------------
