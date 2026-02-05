#!/bin/bash
# Copyright 2022-2026 Salesforce, Inc.
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
set -euxo pipefail

# Lint
bash -n "$(dirname "$0")/install.sh"

# Clean up
rm -rf ~/.slack
if command -v slack-test >/dev/null 2>&1; then
        echo "Error: slack-test is installed"
        exit 1
fi

# Run install script
bash "$(dirname "$0")/install.sh" slack-test
if slack-test --version | grep -q 'v[0-9]\+\.[0-9]\+\.[0-9]\+'; then
        echo "Version found"
else
        echo "No version found"
fi

# Clean up
rm -rf ~/.slack
if slack-test --version 2>/dev/null; then
        echo "Error: The slack-test command was not removed"
        exit 1
fi

# Run install script with version flag
bash "$(dirname "$0")/install.sh" -v 2.2.0 slack-test
if [ "$(slack-test --version)" != "Using slack-test v2.2.0" ]; then
        echo "Error: Version output does not match (as expected)"
        exit 1
fi

# Clean up
rm -rf ~/.slack
if slack-test --version 2>/dev/null; then
        echo "Error: The slack-test command was not removed"
        exit 1
fi
