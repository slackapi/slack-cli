#!/bin/bash
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

# USAGE:
#   ./scripts/archive.sh <artifact path> <release version>
#
# EXAMPLES:
#   Artifacts are built with GoReleaser and must exist before running this
#   script:
#
#   $ make build-snapshot
#   $ ./scripts/archive.sh ./dist 3.3.0
#
# DESCRIPTION:
#   Create the development and production tar.gz and zip archives. These are
#   used in the `install.sh` and `install-dev.sh` scripts.
#
#   Tests for correct packaging is done with the `archive-test.sh` script.

DIST_DIR=${1}

main() {
    if [ $# -lt 2 ]; then
        echo "Missing parameters: $0 <path> <version>"
        exit 1
    fi

    VERSION=${2}

    echo "Creating macOS archives"

    macos_targz_file_path_version_universal="$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.tar.gz"
    macos_targz_file_path_version_amd64="$DIST_DIR/slack_cli_${VERSION}_macOS_amd64.tar.gz"
    macos_targz_file_path_version_arm64="$DIST_DIR/slack_cli_${VERSION}_macOS_arm64.tar.gz"
    macos_targz_file_path_dev_universal="$DIST_DIR/slack_cli_dev_macOS_64-bit.tar.gz"
    macos_targz_file_path_dev_amd64="$DIST_DIR/slack_cli_dev_macOS_amd64.tar.gz"
    macos_targz_file_path_dev_arm64="$DIST_DIR/slack_cli_dev_macOS_arm64.tar.gz"
    macos_targz_file_path_latest_universal="$DIST_DIR/slack_cli_latest_macOS_64-bit.tar.gz"
    macos_targz_file_path_latest_amd64="$DIST_DIR/slack_cli_latest_macOS_amd64.tar.gz"
    macos_targz_file_path_latest_arm64="$DIST_DIR/slack_cli_latest_macOS_arm64.tar.gz"

    macos_zip_file_path_universal="$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.zip"
    macos_zip_file_path_universal_legacy="$DIST_DIR/slack_cli_${VERSION}_macos.zip"

    echo "-> Creating macOS development tar.gz files"
    cp "$macos_targz_file_path_version_universal" "$macos_targz_file_path_dev_universal"
    cp "$macos_targz_file_path_version_amd64" "$macos_targz_file_path_dev_amd64"
    cp "$macos_targz_file_path_version_arm64" "$macos_targz_file_path_dev_arm64"
    ls -l "$DIST_DIR"/*dev_macOS*

    echo "-> Creating macOS latest tar.gz files"
    cp "$macos_targz_file_path_version_universal" "$macos_targz_file_path_latest_universal"
    cp "$macos_targz_file_path_version_amd64" "$macos_targz_file_path_latest_amd64"
    cp "$macos_targz_file_path_version_arm64" "$macos_targz_file_path_latest_arm64"
    ls -l "$DIST_DIR"/*latest_macOS*

    echo "-> Creating macOS legacy .zip file for auto-update"
    cp "$macos_zip_file_path_universal" "$macos_zip_file_path_universal_legacy"
    ls -l "$DIST_DIR"/*_macos.zip

    echo "Creating Linux archives"

    linux_targz_file_path_version="$DIST_DIR/slack_cli_${VERSION}_linux_64-bit.tar.gz"
    linux_targz_file_path_dev="$DIST_DIR/slack_cli_dev_linux_64-bit.tar.gz"
    linux_targz_file_path_latest="$DIST_DIR/slack_cli_latest_linux_64-bit.tar.gz"

    echo "-> Creating Linux development tar.gz file"
    cp "$linux_targz_file_path_version" "$linux_targz_file_path_dev"
    ls -l "$DIST_DIR"/*dev_linux*

    echo "-> Creating Linux production tar.gz file"
    cp "$linux_targz_file_path_version" "$linux_targz_file_path_latest"
    ls -l "$DIST_DIR"/*latest_linux*
}

main "$@"
