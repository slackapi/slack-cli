# Copyright 2022-2025 Salesforce, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Common variables
LDFLAGS=-X 'github.com/slackapi/slack-cli/internal/pkg/version.Version=`git describe --tags --match 'v*.*.*'`'
BUILD_VERSION = `git describe --tags --match 'v*.*.*' | sed 's/v//'`
RELEASE_VERSION := $(shell git describe --tags --match 'v*.*.*')
testdir ?= ...
testname ?= ./...

FONT_BOLD := $(shell tput bold)
FONT_RESET := $(shell tput sgr0)

# Remove files
.PHONY: clean
clean:
	rm -rf ./bin/
	rm -rf ./dist/

# Initialize project
.PHONY: init
init:
	git fetch origin --tags

# Run all unit tests
.PHONY: test
test: build
	go test -ldflags="$(LDFLAGS)" -v ./$(testdir) -run $(testname) -race -covermode=atomic -coverprofile=coverage.out

# Run installation script tests
.PHONY: test-install
test-install: clean
	bash scripts/install-test.sh
	bash scripts/install-dev-test.sh

# Report test coverage
.PHONY: coverage
coverage:
	go tool cover -html=coverage.out

# Run the linter
.PHONY: lint
lint:
	golangci-lint run

# Build the CLI
.PHONY: build
build: lint clean
	mkdir bin/
	go build -ldflags="$(LDFLAGS)" -o bin/slack
	SLACK_DISABLE_TELEMETRY="true" ./bin/slack version --skip-update

# Build the CLI for CI environments. Don't run lint or test
.PHONY: build-ci
build-ci: clean
	mkdir bin/
	go build -ldflags="-s -w $(LDFLAGS)" -o bin/slack
	SLACK_DISABLE_TELEMETRY="true" ./bin/slack version --skip-update

# Build the CLI as a release snapshot for all operating systems
.PHONY: build-snapshot
build-snapshot: clean
	BUILD_VERSION="$(BUILD_VERSION)" LDFLAGS="$(LDFLAGS)" goreleaser --snapshot --clean --skip=publish --config .goreleaser.yml

# Update documentation in a commit tagged as the release
# Usage: `make tag RELEASE_VERSION=3.7.0-example`
.PHONY: tag
tag:
	git diff --quiet --cached
	git diff --quiet docs/guides/installing-the-slack-cli-*.md
	@if echo "$(RELEASE_VERSION)" | grep -q '^v'; then \
		echo "Error: Release version should not begin with a version prefix."; \
		exit 1; \
	fi
	@printf "$(FONT_BOLD)Updating Docs$(FONT_RESET)\n"
	sed -i.bak -E "s#slack_cli_[0-9]+\.[0-9]+\.[0-9]+_macOS_arm64\.tar\.gz#slack_cli_$(RELEASE_VERSION)_macOS_arm64.tar.gz#" docs/guides/installing-the-slack-cli-for-mac-and-linux.md
	sed -i.bak -E "s#slack_cli_[0-9]+\.[0-9]+\.[0-9]+_macOS_amd64\.tar\.gz#slack_cli_$(RELEASE_VERSION)_macOS_amd64.tar.gz#" docs/guides/installing-the-slack-cli-for-mac-and-linux.md
	sed -i.bak -E "s#slack_cli_[0-9]+\.[0-9]+\.[0-9]+_linux_64-bit\.tar\.gz#slack_cli_$(RELEASE_VERSION)_linux_64-bit.tar.gz#" docs/guides/installing-the-slack-cli-for-mac-and-linux.md
	sed -i.bak -E "s#slack_cli_[0-9]+\.[0-9]+\.[0-9]+_windows_64-bit\.zip#slack_cli_$(RELEASE_VERSION)_windows_64-bit.zip#" docs/guides/installing-the-slack-cli-for-windows.md
	sed -i.bak -E "s/Using slack v[0-9]+\.[0-9]+\.[0-9]+/Using slack v$(RELEASE_VERSION)/" docs/guides/installing-the-slack-cli-for-mac-and-linux.md
	sed -i.bak -E "s/Using slack v[0-9]+\.[0-9]+\.[0-9]+/Using slack v$(RELEASE_VERSION)/" docs/guides/installing-the-slack-cli-for-windows.md
	@printf "$(FONT_BOLD)Removing Backups$(FONT_RESET)\n"
	rm docs/guides/installing-the-slack-cli-for-mac-and-linux.md.bak
	rm docs/guides/installing-the-slack-cli-for-windows.md.bak
	git add docs/guides/installing-the-slack-cli-for-mac-and-linux.md
	git add docs/guides/installing-the-slack-cli-for-windows.md
	@printf "$(FONT_BOLD)Git Commit$(FONT_RESET)\n"
	git commit -m "chore: release slack-cli v$(RELEASE_VERSION)"
	@printf "$(FONT_BOLD)Git Tag$(FONT_RESET)\n"
	git tag v$(RELEASE_VERSION)
