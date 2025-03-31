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
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/slackapi/slack-cli/cmd/feedback"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

//go:embed doctor.tmpl
var embedDocTmpl []byte

// DoctorHookJSON holds expected values from the doctor hook
type DoctorHookJSON struct {
	Versions []struct {
		Name    string `json:"name"`
		Current string `json:"current"`
		Message string `json:"message"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"versions"`
}

func NewDoctorCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check and report on system and app information",
		Long: strings.Join([]string{
			"Check and report on relevant system (and sometimes app) dependencies",
			"",
			"System dependencies can be reviewed from any directory",
			"* This includes operating system information and Deno and Git versions",
			"",
			"While app dependencies are only shown within a project directory",
			"* This includes the Deno Slack SDK, API, and hooks versions of an app",
			"* New versions will be listed if there are any updates available",
			"",
			"Unfortunately, the doctor command cannot heal all problems",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "doctor", Meaning: "Create a status report of system dependencies"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			report, err := performChecks(ctx, clients)
			if err != nil {
				return err
			}
			err = style.PrintTemplate(cmd.OutOrStdout(), string(embedDocTmpl), report)
			if err != nil {
				return err
			}
			err = feedback.ShowSurveyMessages(ctx, clients)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// DoctorReport contains information about system statistics
type DoctorReport struct {
	Sections []Section
}

// TotalErrors returns the sum of all errors across sections
func (d DoctorReport) TotalErrors() int {
	totalErrors := 0
	for _, section := range d.Sections {
		totalErrors += section.SumErrors()
	}
	return totalErrors
}

// performChecks runs a series of checks for relevant dependencies.

// If successful, a report containing the details of each
// dependency (and associated errors, if present) is returned
func performChecks(ctx context.Context, clients *shared.ClientFactory) (DoctorReport, error) {
	osSubsection := checkOS(ctx, clients)
	projConfigSubsection := checkProjectConfig(ctx, clients)
	projToolingSubsection := checkProjectTooling(ctx, clients)
	projDepsSubsection := checkProjectDeps(ctx, clients)
	cliSubsection, err := checkCLIVersion(ctx, clients)
	if err != nil {
		return DoctorReport{}, err
	}
	configSubsection, err := checkCLIConfig(ctx, clients)
	if err != nil {
		return DoctorReport{}, err
	}
	credSubsection, err := checkCLICreds(ctx, clients)
	if err != nil {
		return DoctorReport{}, err
	}
	gitSection, err := CheckGit(ctx)
	if err != nil {
		return DoctorReport{}, err
	}

	reportSections := []Section{
		{
			Label: "SYSTEM",
			Subsections: []Section{
				osSubsection,
				gitSection,
			},
		},
		{
			Label: "SLACK",
			Subsections: []Section{
				cliSubsection,
				configSubsection,
				credSubsection,
			},
		},
	}

	if clients.SDKConfig.WorkingDirectory != "" {
		reportSections = append(reportSections, Section{
			Label: "PROJECT",
			Subsections: []Section{
				projConfigSubsection,
				projToolingSubsection,
				projDepsSubsection,
			},
		})
	}

	report := DoctorReport{Sections: reportSections}

	return report, nil
}

// doctorHook returns the response from the doctor hook
func doctorHook(ctx context.Context, clients *shared.ClientFactory) (DoctorHookJSON, error) {
	var doctorHookJSON DoctorHookJSON
	if !clients.SDKConfig.Hooks.Doctor.IsAvailable() {
		return DoctorHookJSON{}, slackerror.New(slackerror.ErrSDKHookNotFound).
			WithMessage("The `doctor` hook was not found").
			WithRemediation("Debug responses from the Slack hooks file (%s)", config.GetProjectHooksJSONFilePath())
	}
	var hookExecOpts = hooks.HookExecOpts{
		Hook: clients.SDKConfig.Hooks.Doctor,
	}
	getDoctorHookJSON, err := clients.HookExecutor.Execute(hookExecOpts)
	if err != nil {
		return DoctorHookJSON{}, err
	}
	err = json.Unmarshal([]byte(getDoctorHookJSON), &doctorHookJSON)
	if err != nil {
		return DoctorHookJSON{}, slackerror.New(slackerror.ErrUnableToParseJson).
			WithMessage("Failed to parse response from doctor hook").
			WithRootCause(err)
	}
	return doctorHookJSON, nil
}
