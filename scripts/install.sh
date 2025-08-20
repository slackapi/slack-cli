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

SLACK_CLI_NAME="slack"
FINGERPRINT="d41d8cd98f00b204e9800998ecf8427e"
SLACK_CLI_VERSION=
rx='^([0-9]+\.){2}(\*|[0-9]+)(-.*)?$'

while getopts "v:d" flag; do
        case "$flag" in
        v)
                if [[ $OPTARG =~ $rx ]]; then
                        SLACK_CLI_VERSION=$OPTARG
                else
                        echo "Slack CLI requires a valid semver version number." >&2
                        return 1
                fi
                ;;
        *)
                >&2 echo -e "\x1b[1m⚠️  Warning: An unknown flag '$1' was passed to the Slack CLI installation script!\x1b[0m"
                ;;
        esac
done

if [ $(($# - $OPTIND)) -lt 1 ]; then
        if [ ! -z ${@:$OPTIND:1} ]; then
                SLACK_CLI_NAME=${@:$OPTIND:1}
        fi
fi

# Print a message then pause for an amount of time
delay() {
        local options="-e"

        # Prevent trailing newlines if "-n" is included
        if [[ "$3" == "-n" ]]; then
                options="-en"
        fi

        echo $options "$2"
        sleep $1
}

# Replace /home/username/folder/file with ~/folder/file
home_path() {
        local input_string="$1"
        echo "${input_string//$HOME/~}"
}

# Originally from https://gist.github.com/jonlabelle/6691d740f404b9736116c22195a8d706
# Echos the inputs, breaks them into separate lines, then sort by semver descending,
# then takes the first line. If that is not the first param, that means $1 < $2
version_lt() {
        test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" != "$1"
}

install_slack_cli() {
        delay 0.6 "🥁 Hello and welcome! Now beginning to install the..."

        delay 0.1 "\x1b[1m      ________ _     _    _____ _    __    _____ _    ________\x1b[0m"
        delay 0.1 "\x1b[1m     /  ______/ |   / \ /  ____/ | /  /  /  ____/ | /___   __/\x1b[0m"
        delay 0.1 "\x1b[1m    /______  |  |  / _ \  |   |      /   | |   |  |    |  |   \x1b[0m"
        delay 0.1 "\x1b[1m     ____ /  |  |___ __ \ |____  |\  \   | |____  |__ _|  |___\x1b[0m"
        delay 0.1 "\x1b[1m   /_______ /|______/  \_\ ____/_| \__\    _____/______/_____/\x1b[0m"
        delay 0.2 "\x1b[0m"

        # Check if slack binary is already in user's system
        if [ -x "$(command -v "$SLACK_CLI_NAME")" ]; then
                delay 0.3 "🔍 Checking if \`$SLACK_CLI_NAME\` already exists on this system..."
                delay 0.2 "⚠️  Heads up! A binary called \`$SLACK_CLI_NAME\` was found!"
                delay 0.3 "🔍 Now checking if it's the same Slack CLI..."

                # Check if command is used for Slack CLI, for Slack CLI with version >= 1.18.0, the fingerprint needs to be matched to proceed installation
                if [[ ! $($SLACK_CLI_NAME _fingerprint) == "$FINGERPRINT" ]] &>/dev/null; then

                        # For Slack CLI with version < 1.18.0, we check with `slack --version` for backwards compatibility
                        if [[ ! $($SLACK_CLI_NAME --version) == *"Using $SLACK_CLI_NAME v"* ]]; then
                                echo -e "🛑 Error: Your existing \`$SLACK_CLI_NAME\` command is different from this Slack CLI!"
                                echo -e "🛑 Halting the install to avoid accidentally overwriting it.\n"

                                echo -e "🔖 Try using an alias when installing to avoid name conflicts:\n"

                                echo -e "curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s your-preferred-alias"
                                return 1
                        fi
                else
                        if [ -z "$SLACK_CLI_VERSION" ]; then
                                delay 0.3 "🍀 It is the same Slack CLI! Upgrading to the latest version..."
                        else
                                delay 0.3 "🍀 It is the same Slack CLI! Switching over to v$SLACK_CLI_VERSION..."
                        fi
                fi
        fi

        if [ -z "$SLACK_CLI_VERSION" ]; then
                #
                # Get the latest published Slack CLI release, the latest release is the most recent non-prerelease, non-draft release, sorted by the created_at attribute.
                # Using grep and sed to parse the semver (excluding "v" to ensure consistence of binaries' filenames ) instead of jq to avoid extra dependencies requirement
                #
                echo -e "🔍 Searching for the latest version of the Slack CLI..."
                LATEST_SLACK_CLI_VERSION=$(curl --silent "https://api.slack.com/slackcli/metadata.json" | grep -o '"version": "[^"]*' | grep -o '[^"]*$' | head -1)
                if [ -z "$LATEST_SLACK_CLI_VERSION" ]; then
                        echo "🛑 Error: Installer cannot find the latest Slack CLI version!"
                        echo "🔖 Check the status of https://slack-status.com/ and try again"
                        return 1
                fi
                echo -e "💾 Release v$LATEST_SLACK_CLI_VERSION was found! Downloading now..."
                SLACK_CLI_VERSION=$LATEST_SLACK_CLI_VERSION
        fi

        #
        # Install Slack CLI
        #

        if [ "$(uname)" == "Darwin" ]; then
                if version_lt "$SLACK_CLI_VERSION" "3.3.0"; then
                        slack_cli_url="https://downloads.slack-edge.com/slack-cli/slack_cli_${SLACK_CLI_VERSION}_macOS_64-bit.tar.gz"
                else
                        case "$(uname -m)" in
                        x86_64)
                                slack_cli_url="https://downloads.slack-edge.com/slack-cli/slack_cli_${SLACK_CLI_VERSION}_macOS_amd64.tar.gz"
                                ;;
                        arm64 | aarch64)
                                slack_cli_url="https://downloads.slack-edge.com/slack-cli/slack_cli_${SLACK_CLI_VERSION}_macOS_arm64.tar.gz"
                                ;;
                        *)
                                slack_cli_url="https://downloads.slack-edge.com/slack-cli/slack_cli_${SLACK_CLI_VERSION}_macOS_64-bit.tar.gz"
                                ;;
                        esac
                fi
        elif [ "$(expr substr "$(uname -s)" 1 5)" == "Linux" ]; then
                slack_cli_url="https://downloads.slack-edge.com/slack-cli/slack_cli_${SLACK_CLI_VERSION}_linux_64-bit.tar.gz"
        else
                echo "🛑 Error: This installer is only supported on Linux and macOS"
                echo "🔖 Try using a different installation method:"
                echo "🔗 https://docs.slack.dev/tools/slack-cli"
                return 1
        fi

        slack_cli_install_dir="$HOME/.slack"
        slack_cli_install_bin_dir="$slack_cli_install_dir/bin"
        slack_cli_bin_path="$slack_cli_install_bin_dir/slack"

        if [ ! -d "$slack_cli_install_dir" ]; then
                mkdir -p "$slack_cli_install_dir"
        fi

        echo -e "\x1b[2m\n$slack_cli_url"
        curl -# -fLo "$slack_cli_install_dir/slack-cli.tar.gz" "$slack_cli_url"
        echo -e "\x1b[0m"
        delay 0.2 "💾 Successfully downloaded Slack CLI v$LATEST_SLACK_CLI_VERSION to $(home_path "$slack_cli_install_dir/slack-cli.tar.gz")"

        delay 0.3 "📦 Extracting the Slack CLI command binary to $(home_path "$slack_cli_bin_path")"
        tar -xf "$slack_cli_install_dir/slack-cli.tar.gz" -C "$slack_cli_install_dir"
        chmod +x "$slack_cli_bin_path"

        delay 0.1 "📠 Removing packaged download files from $(home_path "$slack_cli_install_dir/slack-cli.tar.gz")"
        rm "$slack_cli_install_dir/slack-cli.tar.gz"

        delay 0.1 "🔗 Adding a symbolic link from /usr/local/bin/$SLACK_CLI_NAME to $(home_path "$slack_cli_bin_path")"
        if [ ! -d /usr/local/bin ]; then
                echo -e "⚠️  The /usr/local/bin directory does not exist!"
                echo -e "🔐 Please create /usr/local/bin directory first and try again..."
                return 1
        fi
        if [ -w /usr/local/bin ]; then
                ln -sf "$slack_cli_bin_path" "/usr/local/bin/$SLACK_CLI_NAME"
        else
                echo -e "⚠️  Failed to create a symbolic link!"
                delay 0.1 "🔖 The installer doesn't have write access to /usr/local/bin"
                echo -e "🔐 Please check permission and try again..."
                return 1
        fi
        if ! command -v "$SLACK_CLI_NAME" >/dev/null 2>&1; then
                echo "📁 Manually add the Slack CLI command directory to your shell profile"
                echo "   export PATH=\"$slack_cli_install_bin_dir:\$PATH\""
        fi
}

terms_of_service() {
        echo -e ""
        echo -e "📄 Use of the Slack CLI should comply with the Slack API Terms of Service:"
        echo -e "🏛️  https://slack.com/terms-of-service/api"
}

feedback_message() {
        CODE=$?
        echo -e "\x1b[0m"
        echo -e "💌 We would love to know how things are going. Really. All of it."
        if command -v "$SLACK_CLI_NAME" >/dev/null 2>&1; then
                echo -e "✨ Survey your development experience with \`$SLACK_CLI_NAME feedback\`"
        else
                echo -e "✨ Submit installation issues: https://github.com/slackapi/slack-cli/issues"
        fi
        if [ $CODE -ne 0 ]; then
                exit $CODE
        fi
}

next_step_message() {
        echo -e ""
        if command -v "$SLACK_CLI_NAME" >/dev/null 2>&1; then
                echo -e "📺 Success! The Slack CLI is now installed!"
                echo -e "🔐 Next, authorize your CLI in your workspace with \`$SLACK_CLI_NAME login\`"
        fi
}

main() {
        trap 'feedback_message' ERR

        set -eE
        install_slack_cli "$@"
        sleep 0.2
        terms_of_service
        sleep 0.1
        feedback_message
        sleep 0.2
        next_step_message
}

main "$@"
