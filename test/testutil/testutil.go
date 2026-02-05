// Copyright 2022-2026 Salesforce, Inc.
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

package testutil

import (
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

// package + .test for root command
var rootName string = "cmd.test"

// ContainsSemVer checks if a string contains valid semver
func ContainsSemVer(s string) bool {
	matcher := regexp.MustCompile(version.SemverRegexpRaw)
	match := matcher.MatchString(s)
	return match
}

// Set the command's IOStream to the mocked IOStream
// Outside of testing, this is done in the root command's PersistentPreRun function
func MockCmdIO(io iostreams.IOStreamer, cmd *cobra.Command) {
	io.SetCmdIO(cmd)

	if cmd.Name() == rootName {
		fakeRunFunc := func(c *cobra.Command, args []string) error { return nil }
		cmd.PersistentPostRunE = fakeRunFunc
		cmd.PersistentPreRunE = fakeRunFunc
	}
}
