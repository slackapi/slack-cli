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
