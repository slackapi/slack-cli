#!/usr/bin/env bash
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
set -e

#
# USAGE:
#   ./scripts/archive-test.sh <release version>
#
# EXAMPLES:
#   Artifacts are built with GoReleaser and should be packaged into various
#   release archives before running this script:
#
#   $ make build-snapshot
#   $ ./scripts/archive.sh ./dist 3.3.0
#   $ ./scripts/archive-test.sh ./dist 3.3.0
#
# DESCRIPTION:
#   Confirm the expected tar.gz and zip bundles exist for a packaged version.
#
#   Various binaries are extracted and checked for existence and permissions.

DIST_DIR=${1}

get_version_major() {
    local version="$1"
    echo "${version}" | cut -d'.' -f1
}

get_version_major_minor() {
    local version="$1"
    echo "${version}" | cut -d'.' -f1-2
}

main() {
    if [ $# -lt 2 ]; then
        echo "Missing parameters: $0 <path> <version>"
        exit 1
    fi

    VERSION=${2}
    VERSION_MAJOR=$(get_version_major "$VERSION")
    VERSION_MAJOR_MINOR=$(get_version_major_minor "$VERSION")

    echo "Checking macOS archives"
    check_tar "$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION}_macOS_amd64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION}_macOS_arm64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR}.x.x_macOS_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR}.x.x_macOS_amd64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR}.x.x_macOS_arm64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR_MINOR}.x_macOS_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR_MINOR}.x_macOS_amd64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR_MINOR}.x_macOS_arm64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_dev_macOS_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_dev_macOS_amd64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_dev_macOS_arm64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_latest_macOS_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_latest_macOS_amd64.tar.gz"
    check_tar "$DIST_DIR/slack_cli_latest_macOS_arm64.tar.gz"

    check_zip "$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.zip"
    check_zip "$DIST_DIR/slack_cli_${VERSION}_macOS_amd64.zip"
    check_zip "$DIST_DIR/slack_cli_${VERSION}_macOS_arm64.zip"
    check_zip "$DIST_DIR/slack_cli_${VERSION}_macos.zip"

    echo "Checking Linux archives"
    check_tar "$DIST_DIR/slack_cli_${VERSION}_linux_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR}.x.x_linux_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_${VERSION_MAJOR_MINOR}.x_linux_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_dev_linux_64-bit.tar.gz"
    check_tar "$DIST_DIR/slack_cli_latest_linux_64-bit.tar.gz"

    echo "Checking Windows archives"
    check_exe "$DIST_DIR/slack_cli_${VERSION}_windows_64-bit.zip"
    check_exe "$DIST_DIR/slack_cli_${VERSION_MAJOR}.x.x_windows_64-bit.zip"
    check_exe "$DIST_DIR/slack_cli_${VERSION_MAJOR_MINOR}.x_windows_64-bit.zip"
    check_exe "$DIST_DIR/slack_cli_dev_windows_64-bit.zip"
    check_exe "$DIST_DIR/slack_cli_latest_windows_64-bit.zip"

    echo "Success! Archives exist!"
}

check_exe() {
    echo "-> Testing executable exists: $1"
    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT
    unzip -q "$1" -d "$tmpdir"

    file "$tmpdir/bin/slack.exe" | grep -q 'PE32' || {
        echo "-> Failed to find executable: $1"
        exit 1
    }
}

check_tar() {
    echo "-> Testing executable exists: $1"
    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT
    tar -xzf "$1" -C "$tmpdir"

    if ! [[ -x "$tmpdir/bin/slack" ]]; then
        echo "-> Failed to find executable: $1"
        return 1
    fi
}

check_zip() {
    echo "-> Testing executable exists: $1"
    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT
    unzip -q "$1" -d "$tmpdir"

    if ! [[ -x "$tmpdir/slack" ]]; then
        echo "-> Failed to find executable: $1"
        return 1
    fi
}

main "$@"
