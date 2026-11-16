#!/usr/bin/env bash
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

#
# USAGE:
#   ./scripts/triage-milestone-test.sh
#
# EXAMPLES:
#   $ ./scripts/triage-milestone-test.sh
#
# DESCRIPTION:
#   Confirm the milestone triage script rejects unexpected release tags.
#
#   Only the checks that happen before a GitHub API call are covered, so no
#   credentials or network access are needed.

TRIAGE_SCRIPT="$(dirname "$0")/triage-milestone.sh"

# Lint
bash -n "$TRIAGE_SCRIPT"

# Missing release tag
if bash "$TRIAGE_SCRIPT" >/dev/null 2>&1; then
    echo "Error: A missing release tag should exit with an error"
    exit 1
fi

# Release tags of an unexpected format
for tag in "4.8.0" "v4.8" "v4.8.0.1" "v4.8.0-example-feature" "dev-build" ""; do
    if bash "$TRIAGE_SCRIPT" "$tag" >/dev/null 2>&1; then
        echo "Error: The '$tag' release tag should exit with an error"
        exit 1
    fi
done
