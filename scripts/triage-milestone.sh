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
set -euo pipefail

#
# USAGE:
#   ./scripts/triage-milestone.sh <release tag>
#
# EXAMPLES:
#   Triage the milestone after the v4.8.0 release is published:
#
#   $ ./scripts/triage-milestone.sh v4.8.0
#
#   Print the changes without making them:
#
#   $ DRY_RUN=true ./scripts/triage-milestone.sh v4.8.0
#
#   Rehearse on a scratch repository before a real release:
#
#   $ REPO=example/scratch ./scripts/triage-milestone.sh v0.1.0
#
# DESCRIPTION:
#   Roll the "Next Release" milestone over to a published release tag.
#
#   The "Next Release" milestone is renamed to the release tag and closed while
#   a new "Next Release" milestone collects the issues and pull requests that
#   remain open. Merged and closed items stay on the release tag milestone and
#   items without a milestone are left alone.
#
#   The gh command is required, along with a GH_TOKEN that can write issues.
#
#   Completed steps are skipped, so an interrupted run is repeated safely.

REPO=${REPO:-slackapi/slack-cli}
DRY_RUN=${DRY_RUN:-false}
UPCOMING_TITLE="Next Release"

main() {
    if [ $# -lt 1 ]; then
        echo "Missing parameters: $0 <release tag>"
        exit 1
    fi

    TAG=${1}

    if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: The '$TAG' tag is not a production release tag of vX.Y.Z"
        exit 1
    fi

    if ! command -v gh >/dev/null 2>&1; then
        echo "Error: The gh command is required but was not found"
        exit 1
    fi

    echo "Triaging the \"$UPCOMING_TITLE\" milestone of $REPO for $TAG"

    # The release tag milestone might exist if an earlier run was interrupted
    released=$(find_milestone "$TAG" "all")
    upcoming=""

    if [ -z "$released" ]; then
        released=$(find_milestone "$UPCOMING_TITLE" "open")
        if [ -z "$released" ]; then
            echo "-> No \"$UPCOMING_TITLE\" milestone was found, so nothing is triaged"
            exit 0
        fi
        echo "-> Renaming milestone #$released to $TAG"
        edit_milestone "$released" "title=$TAG"
    else
        echo "-> Milestone $TAG exists as #$released"
        upcoming=$(find_milestone "$UPCOMING_TITLE" "open")
    fi

    if [ -z "$upcoming" ]; then
        echo "-> Creating the next \"$UPCOMING_TITLE\" milestone"
        upcoming=$(create_milestone "$UPCOMING_TITLE")
    else
        echo "-> Milestone \"$UPCOMING_TITLE\" exists as #$upcoming"
    fi

    echo "-> Moving open issues and pull requests to \"$UPCOMING_TITLE\""
    move_open_issues "$released" "$upcoming"

    echo "-> Closing milestone $TAG"
    edit_milestone "$released" "state=closed"

    echo "Triaged the milestone of $TAG"
}

# Check if changes should be printed instead of made
is_dry_run() {
    [ "$DRY_RUN" = "true" ]
}

# Output the number of the milestone matching a title and state, if one exists
find_milestone() {
    local numbers
    numbers=$(gh api "repos/$REPO/milestones?state=${2}&per_page=100" --paginate \
        --jq ".[] | select(.title == \"${1}\") | .number")
    echo "$numbers" | head -n 1
}

# Create a milestone with a title and output the new milestone number
create_milestone() {
    if is_dry_run; then
        echo "0"
        return
    fi
    gh api --method POST "repos/$REPO/milestones" -f "title=${1}" --jq ".number"
}

# Change a single field of an existing milestone
edit_milestone() {
    if is_dry_run; then
        echo "   Skipping the '${2}' change of milestone #${1}"
        return
    fi
    gh api --method PATCH "repos/$REPO/milestones/${1}" -f "${2}" --silent
}

# Assign the open issues and pull requests of a milestone to another milestone
#
# Both issues and pull requests are returned by the issues endpoint, which is
# preferred over a search because search results are not immediately current.
move_open_issues() {
    local number
    for number in $(gh api "repos/$REPO/issues?milestone=${1}&state=open&per_page=100" \
        --paginate --jq ".[].number"); do
        if is_dry_run; then
            echo "   Skipping the milestone change of #$number"
            continue
        fi
        echo "   #$number"
        gh api --method PATCH "repos/$REPO/issues/$number" -F "milestone=${2}" --silent
    done
}

main "$@"
