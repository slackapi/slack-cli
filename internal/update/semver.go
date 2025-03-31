// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package update

import (
	"github.com/hashicorp/go-version"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// SemVerGreaterThan returns true if release is greater than current
func SemVerGreaterThan(release string, current string) (bool, error) {
	releaseVersion, err := version.NewVersion(release)
	if err != nil {
		return false, slackerror.New(slackerror.ErrInvalidSemVer).WithRootCause(err)
	}
	currentVersion, err := version.NewVersion(current)
	if err != nil {
		return false, slackerror.New(slackerror.ErrInvalidSemVer).WithRootCause(err)
	}
	return releaseVersion.GreaterThan(currentVersion), nil
}

// SemVerLessThan returns true if release is less than current
func SemVerLessThan(release string, current string) (bool, error) {
	releaseVersion, err := version.NewVersion(release)
	if err != nil {
		return false, slackerror.New(slackerror.ErrInvalidSemVer).WithRootCause(err)
	}
	currentVersion, err := version.NewVersion(current)
	if err != nil {
		return false, slackerror.New(slackerror.ErrInvalidSemVer).WithRootCause(err)
	}
	return releaseVersion.LessThan(currentVersion), nil
}
