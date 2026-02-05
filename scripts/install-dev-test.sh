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
bash -n "$(dirname "$0")/install-dev.sh"

# Clean up
rm -rf ~/.slack/dev-build
if command -v slack-dev-test >/dev/null 2>&1; then
        echo "Error: slack-test is installed"
        exit 1
fi

# Run install-dev script
bash "$(dirname "$0")/install-dev.sh" slack-dev-test
if slack-dev-test --version | grep -q 'v[0-9]\+\.[0-9]\+\.[0-9]\+'; then
        echo "Version found"
else
        echo "No version found"
fi

# Clean up
rm -rf ~/.slack/dev-build
if slack-dev-test --version 2>/dev/null; then
        echo "Error: The slack-dev-test command was not removed"
        exit 1
fi
