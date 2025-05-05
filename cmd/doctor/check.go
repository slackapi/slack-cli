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

package doctor

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/deputil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/internal/update"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Section struct {
	Label       string
	Value       string
	Subsections []Section
	Errors      []slackerror.Error
}

// HasError returns if errors exist in any section or subsection
func (s *Section) HasError() bool {
	return s.SumErrors() > 0
}

// SumErrors returns the error count in sections and subsections
func (s *Section) SumErrors() int {
	totalErrors := 0
	if len(s.Errors) > 0 {
		totalErrors += len(s.Errors)
	}
	for _, subsection := range s.Subsections {
		totalErrors += subsection.SumErrors()
	}
	return totalErrors
}

// RenderLabel formats a label for use with optional values
func (s *Section) RenderLabel() string {
	if len(s.Value) > 0 {
		return s.Label + ":"
	}
	return s.Label
}

// checkOS returns the operating system information of the user's system
func checkOS(ctx context.Context, clients *shared.ClientFactory) Section {
	osSection := Section{
		Label: "Operating System",
		Value: osDescription(clients.IO),
	}

	osVersion, osArch := runtime.GOOS, runtime.GOARCH
	versionSection := Section{
		Label: "Version",
		Value: fmt.Sprintf("%s (%s)", osVersion, osArch),
	}

	switch osVersion {
	case "windows":
	case "darwin":
	case "linux":
	default:
		osErr := slackerror.ErrorCodeMap[slackerror.ErrOSNotSupported]
		versionSection.Errors = []slackerror.Error{osErr}
	}

	osSection.Subsections = []Section{versionSection}

	return osSection
}

// osDescription returns a short and random definition of an operating system
func osDescription(io iostreams.IOStreamer) string {
	descriptions := []string{
		"the computer conductor",
		"system management software",
		"a processor of processes",
		"the kernel and drivers",
		"a user-computer interface",
		"the computer command center",
		"virtual machine orchestrator",
		"program scheduler and such",
		"resource allocation manager",
		"the hardware guardian",
		"digital infrastructure controller",
		"hardware access mediator",
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	choice := rand.Intn(len(descriptions))

	// Remove choice and chance in scripting
	if !io.IsTTY() {
		choice = 0
	}
	return descriptions[choice]
}

// checkCLIVersion returns the installed version of the user's Slack CLI
func checkCLIVersion(ctx context.Context, clients *shared.ClientFactory) (Section, error) {
	cliVersion := version.Version
	versionSection := Section{"Version", cliVersion, []Section{}, []slackerror.Error{}}
	return Section{"CLI", "this tool for building Slack apps", []Section{versionSection}, []slackerror.Error{}}, nil
}

// checkProjectConfig returns details about the current project configurations
// or returns an empty section if not in a project directory
func checkProjectConfig(ctx context.Context, clients *shared.ClientFactory) Section {
	section := Section{
		Label: "Configurations",
		Value: "your project's CLI settings",
	}
	projectConfig, err := clients.Config.ProjectConfig.ReadProjectConfigFile(ctx)
	if err != nil {
		if slackerror.ToSlackError(err).Code != slackerror.ErrInvalidAppDirectory {
			section.Errors = append(section.Errors, *slackerror.ToSlackError(err))
		}
		return section
	}
	if projectConfig.Manifest != nil && projectConfig.Manifest.Source != "" {
		section.Subsections = append(section.Subsections, Section{
			Label: "Manifest source",
			Value: projectConfig.Manifest.Source,
		})
	} else {
		section.Errors = append(section.Errors,
			*slackerror.New(slackerror.ErrProjectConfigManifestSource),
		)
	}
	if projectConfig.ProjectID != "" {
		section.Subsections = append(section.Subsections, Section{
			Label: "Project ID",
			Value: projectConfig.ProjectID,
		})
	} else {
		section.Errors = append(section.Errors,
			*slackerror.New(slackerror.ErrProjectConfigIDNotFound),
		)
	}
	return section
}

// checkProjectDeps returns details about the current project's dependencies
func checkProjectDeps(ctx context.Context, clients *shared.ClientFactory) Section {
	section := Section{
		Label: "Dependencies",
		Value: "requisites for development",
	}
	checkUpdateJSON, err := update.CheckUpdateHook(ctx, clients)
	if err != nil {
		slackErr := slackerror.ToSlackError(err)
		if slackErr.Code == slackerror.ErrSDKHookInvocationFailed {
			slackErr.Remediation = "Check that the check-update hook command is valid"
		}
		section.Errors = append(section.Errors, *slackErr)
		return section
	}
	for _, release := range checkUpdateJSON.Releases {
		dependencyText := release.Current
		if release.Update {
			latest := style.CommandText(fmt.Sprintf(`%s (update available)`, release.Latest))
			dependencyText = fmt.Sprintf(`%s → %s`, release.Current, latest)
		} else if release.Current != release.Latest {
			// If there's no update but the versions don't match, warn of unsupported version
			latest := style.CommandText(fmt.Sprintf(`%s (supported version)`, release.Latest))
			dependencyText = fmt.Sprintf(`%s → %s`, release.Current, latest)
		}
		section.Subsections = append(section.Subsections, Section{
			Label: release.Name,
			Value: dependencyText,
		})
	}
	return section
}

// checkCLIConfig reads the contents of config.json
// and outputs details of the configuration settings
func checkCLIConfig(ctx context.Context, clients *shared.ClientFactory) (Section, error) {
	section := Section{"Configurations", "any adjustments to settings", []Section{}, []slackerror.Error{}}

	userConfig, err := clients.Config.SystemConfig.UserConfig(ctx)
	if err != nil {
		return Section{}, slackerror.Wrap(err, "Failed to read system configuration")
	}

	// System ID
	systemIDSubsection := Section{
		"System ID",
		userConfig.SystemID,
		[]Section{},
		[]slackerror.Error{},
	}

	if userConfig.SystemID == "" {
		errSystemID := slackerror.ErrorCodeMap[slackerror.ErrSystemConfigIDNotFound]
		systemIDSubsection.Errors = []slackerror.Error{errSystemID}
	}

	// Last Updated
	lastUpdatedSubsection := Section{
		"Last updated",
		userConfig.LastUpdateCheckedAt.Format("2006-01-02 15:04:05 Z07:00"),
		[]Section{},
		[]slackerror.Error{},
	}

	// Experiments
	allConfigExperiments := "None"
	allExperiments := []string{}
	for _, exp := range clients.Config.GetExperiments() {
		allExperiments = append(allExperiments, string(exp))
	}
	if len(allExperiments) > 0 {
		allConfigExperiments = strings.Join(allExperiments, ", ")
	}

	experimentsSubsection := Section{
		"Experiments",
		allConfigExperiments,
		[]Section{},
		[]slackerror.Error{},
	}

	// Build the list of subsections
	subsection := []Section{
		systemIDSubsection,
		lastUpdatedSubsection,
		experimentsSubsection,
	}

	section.Subsections = subsection

	return section, nil
}

// checkCLICreds reads the contents of credentials.json
// and outputs information for each team listed
func checkCLICreds(ctx context.Context, clients *shared.ClientFactory) (Section, error) {
	section := Section{"Credentials", "your Slack authentication", []Section{}, []slackerror.Error{}}

	authList, err := clients.AuthInterface().Auths(ctx)
	if err != nil {
		return Section{}, slackerror.New(slackerror.ErrAuthToken).WithRootCause(err)
	}

	// No teams
	if len(authList) == 0 {
		section.Errors = []slackerror.Error{*slackerror.New(slackerror.ErrNotAuthed)}
	}

	// Teams
	if len(authList) > 0 {
		authSections := []Section{}
		currentAPIHost := clients.Config.APIHostResolved
		caser := cases.Title(language.English)
		for _, authInfo := range authList {
			checkDetails := []Section{
				{"Team domain", authInfo.TeamDomain, []Section{}, []slackerror.Error{}},
				{"Team ID", authInfo.TeamID, []Section{}, []slackerror.Error{}},
				{"User ID", authInfo.UserID, []Section{}, []slackerror.Error{}},
				{
					"Last updated",
					authInfo.LastUpdated.Format("2006-01-02 15:04:05 Z07:00"),
					[]Section{},
					[]slackerror.Error{},
				},
				{"Authorization level", caser.String(authInfo.AuthLevel()), []Section{}, []slackerror.Error{}},
			}

			if authInfo.APIHost != nil {
				hostSection := Section{"API Host", *authInfo.APIHost, []Section{}, []slackerror.Error{}}
				checkDetails = append(checkDetails, hostSection)
			}

			// Validate session token
			validitySection := Section{"Token status", "Valid", []Section{}, []slackerror.Error{}}

			// TODO :: .ValidateSession() utilizes the host (APIHost) assigned to the client making
			// the call. This results in incorrectly deeming tokens invalid if using multiple workspaces
			// with different API hosts. (cc: @mbrooks)
			clients.Config.APIHostResolved = clients.AuthInterface().ResolveAPIHost(ctx, clients.Config.APIHostFlag, &authInfo)
			_, err := clients.API().ValidateSession(ctx, authInfo.Token)
			if err != nil {
				validitySection.Value = "Invalid"
			}
			checkDetails = append(checkDetails, validitySection)

			authSection := Section{"", "", checkDetails, []slackerror.Error{}}
			authSections = append(authSections, authSection)
		}

		clients.Config.APIHostResolved = currentAPIHost

		section.Subsections = authSections
	}

	return section, nil
}

// checkProjectTooling collects dependencies required for project execution
func checkProjectTooling(ctx context.Context, clients *shared.ClientFactory) Section {
	toolingSections := []Section{}
	toolingErrors := []slackerror.Error{}

	doctorJSON, err := doctorHook(ctx, clients)
	if err != nil {
		slackErr := slackerror.ToSlackError(err)
		if slackErr.Code == slackerror.ErrSDKHookInvocationFailed {
			slackErr.Remediation = "Check that the doctor hook command is valid"
		}
		toolingErrors = append(toolingErrors, *slackerror.ToSlackError(err))
	} else {
		for _, version := range doctorJSON.Versions {
			versionSections := []Section{}
			versionErrors := []slackerror.Error{}
			if version.Message != "" {
				latest := fmt.Sprintf("Note: %s", version.Message)
				versionSections = append(versionSections, Section{
					Label: style.Secondary(latest),
				})
			}
			if version.Error.Message != "" {
				versionErrors = append(
					versionErrors,
					*slackerror.New(slackerror.ErrRuntimeNotSupported).
						WithMessage("%s", version.Error.Message),
				)
			}
			toolingSections = append(toolingSections, Section{
				Label:       version.Name,
				Value:       version.Current,
				Subsections: versionSections,
				Errors:      versionErrors,
			})
		}
	}
	section := Section{
		Label:       "Runtime",
		Value:       "foundations for the application",
		Subsections: toolingSections,
		Errors:      toolingErrors,
	}
	return section
}

// CheckGit checks for the version of an installed Git on the user machine
func CheckGit(ctx context.Context) (Section, error) {
	gitSection := Section{"Git", "a version control system", []Section{}, []slackerror.Error{}}

	version, err := deputil.GetGitVersion()
	if err != nil {
		gitSection.Errors = []slackerror.Error{*slackerror.ToSlackError(err)}
		return gitSection, nil
	}
	versionSection := Section{"Version", string(version), []Section{}, []slackerror.Error{}}

	gitSection.Subsections = []Section{versionSection}
	return gitSection, nil
}
