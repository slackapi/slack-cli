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


#
# USAGE:
#   ./scripts/archive-test.sh <release version>
#
# EXAMPLES:
#   This script now executed inside goreleaser, if you need to run the script alone, make sure you have macOS and Linux original archives in ./dist
#   - ./scripts/archive-test.sh 1.5.0
# DESCRIPTION:
#   Create the development tar.gz archives used by the `install-dev.sh` script.
#
#   Generates macOS and Linux archives.
#
#   After generating the archive, the script will test the archive by
#   extracting it and looking for `bin/slack` for macOS binary
#   Linux archive is generated directorily from original binary

DIST_DIR="dist"

main() {
    if [ $# -lt 1 ]; then
        echo "Release version is required"
        exit 1
    fi

    VERSION=${1}

    echo "Creating macOS archive:"

    src_file_path="$DIST_DIR/slack-macos_darwin_amd64/slack_cli_${VERSION}_macos.zip"
    echo "-> Source archive: $src_file_path"

    targz_dir_path="$DIST_DIR/slack_mac_dev_64-bit"
    targz_file_path="$DIST_DIR/slack_mac_dev_64-bit.tar.gz"

    echo "-> Creating dist directory: $targz_dir_path"
    mkdir -p "$targz_dir_path"

    echo "-> Extracting: $targz_dir_path"
    unzip "$src_file_path" -d "$targz_dir_path"

    echo "-> Moving $targz_dir_path/slack to $targz_dir_path/bin/slack"
    mkdir "$targz_dir_path/bin"
    mv "$targz_dir_path/slack" "$targz_dir_path/bin/slack"

    echo "-> Creating tar.gz file: $targz_file_path"
    tar -C "$targz_dir_path" -zcvf "$targz_file_path" .

    echo "-> Cleaning up: $targz_dir_path"
    rm -R "$targz_dir_path"

    echo "-> Testing tar.gz file..."
    mkdir "$targz_dir_path"
    tar -zxvf "$targz_file_path" -C "$targz_dir_path"

    if [ $(command -v "$targz_dir_path/bin/slack") ]; then
      echo "-> Found bin/slack"
      rm -R "$targz_dir_path"
    else
      echo "-> Missing bin/slack"
    fi

    echo "-> Everything looks good!"
    echo "-> macOS archive: $targz_file_path"

    echo "Creating Linux archive:"

    linux_src_file_path="$DIST_DIR/slack_linux_amd64/"
    linux_targz_file_path="$DIST_DIR/slack_linux_dev_64-bit.tar.gz"

    if [ $(command -v "$linux_src_file_path/bin/slack") ]; then
      echo "-> Found bin/slack"
    else
      echo "-> Missing bin/slack"
      exit
    fi

    echo "-> Creating tar.gz file: $targz_file_path"
    tar -czvf "$linux_targz_file_path" -C "$linux_src_file_path" .

    echo "-> Linux archive: $linux_targz_file_path"
}

main "$@"
