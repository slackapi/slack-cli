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
set -eo

# USAGE:
#   ./scripts/archive.sh <artifact path> <release version>
#
# EXAMPLES:
#   This script now executed inside goreleaser, if you need to run the script alone, make sure you have macOS and Linux original archives in your artificate path
#   - ./scripts/archive.sh /artifacts 1.5.0
#
# DESCRIPTION:
#   Create the development and production tar.gz archives used by the `install.sh` and `install-dev.sh`scripts.
#
#   Generates macOS and Linux archives.

#   After generating the archive for macOS, the script will test the archive by
#   extracting it and looking for `bin/slack`
#   Linux archive is generated directorily from original binary,
#   the  script will check if `bin/slack` for Linux exits or not first

DIST_DIR=${1}

main() {
    if [ $# -lt 2 ]; then
        echo "Missing parameters"
        exit 1
    fi

    VERSION=${2}

    echo "Creating macOS archives:"

    mac_src_file_path="$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.zip"
    echo "-> Source archive: $mac_src_file_path"

    mac_targz_dir_path="$DIST_DIR/slack_mac_64-bit"
    macos_dev_targz_file_path="$DIST_DIR/slack_cli_dev_macOS_64-bit.tar.gz"
    macos_prod_targz_file_path="$DIST_DIR/slack_cli_latest_macOS_64-bit.tar.gz"
    macos_prod_targz_file_path_2="$DIST_DIR/slack_cli_${VERSION}_macOS_64-bit.tar.gz"
    macos_prod_zip_file_path="$DIST_DIR/slack_cli_${VERSION}_macos.zip"

    echo "-> Creating dist directory: $mac_targz_dir_path"
    mkdir -p "$mac_targz_dir_path"

    echo "-> Extracting: $mac_targz_dir_path"
    unzip "$mac_src_file_path" -d "$mac_targz_dir_path"

    echo "-> Moving $mac_targz_dir_path/slack to $mac_targz_dir_path/bin/slack"
    mkdir "$mac_targz_dir_path/bin"
    mv "$mac_targz_dir_path/slack" "$mac_targz_dir_path/bin/slack"

    echo "-> Creating macOS development tar.gz file: $macos_dev_targz_file_path"
    tar -C "$mac_targz_dir_path" -zcvf "$macos_dev_targz_file_path" .

    echo "-> Creating macOS production tar.gz file: $macos_prod_targz_file_path"
    tar -C "$mac_targz_dir_path" -zcvf "$macos_prod_targz_file_path" .
    tar -C "$mac_targz_dir_path" -zcvf "$macos_prod_targz_file_path_2" .

    echo "-> Creating macOS .zip file with old name for auto-update:
    $macos_prod_zip_file_path"
    cp -i "$mac_src_file_path" "$macos_prod_zip_file_path"

    echo "-> Cleaning up: $mac_targz_dir_path"
    rm -R "$mac_targz_dir_path"

    echo "-> Testing macOs development tar.gz file..."
    mkdir "$mac_targz_dir_path"
    tar -zxvf "$macos_dev_targz_file_path" -C "$mac_targz_dir_path"

    if [ $(command -v "$mac_targz_dir_path/bin/slack") ]; then
        echo "-> Found bin/slack"
        rm -R "$mac_targz_dir_path"
    else
        echo "-> Missing development bin/slack"
    fi

    echo "-> Testing macOs production tar.gz file..."
    mkdir "$mac_targz_dir_path"
    tar -zxvf "$macos_prod_targz_file_path" -C "$mac_targz_dir_path"

    if [ $(command -v "$mac_targz_dir_path/bin/slack") ]; then
        echo "-> Found bin/slack"
        rm -R "$mac_targz_dir_path"
    else
        echo "-> Missing production bin/slack"
    fi

    echo "-> Everything looks good!"
    echo "-> macOS archive: $targz_file_path"

    echo "Creating Linux archives:"

    linux_src_file_path="$DIST_DIR/slack_cli_${VERSION}_linux_64-bit.tar.gz"
    linux_dev_targz_file_path="$DIST_DIR/slack_cli_dev_linux_64-bit.tar.gz"
    linux_prod_targz_file_path="$DIST_DIR/slack_cli_latest_linux_64-bit.tar.gz"

    if [ -f "$linux_src_file_path" ]; then
        echo "-> Found linux binary archive"
    else
        echo "-> Missing linux binary archive"
        exit
    fi

    echo "-> Creating development Linux tar.gz file: $linux_dev_targz_file_path"
    cp "$linux_src_file_path" "$linux_dev_targz_file_path"
    echo "-> Linux development archive: $linux_dev_targz_file_path"

    echo "-> Creating production Linux tar.gz file: $linux_prod_targz_file_path"
    cp "$linux_src_file_path" "$linux_prod_targz_file_path"
    echo "-> Linux development archive: $linux_prod_targz_file_path"
}

main "$@"
