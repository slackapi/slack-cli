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

package triggers

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

var defaultTriggerPaths = []string{"triggers/*"}

// File paths matching these regex sequences will be ignored
// In the future this default can be overridden by
// be user-definition within config.json
// ends with .ts~, .ts~1, .ts~123~
var defaultTriggerFilePathMatchers = [...]string{".*.~\\d?"}

// TriggerGenerate prompts the user to create a trigger when no triggers exist for the app provided. The prompt displays a list of
// trigger definition files that are available in the project. Depending on whether the user chooses to create a trigger
// or skip the step, the function returns the created trigger.
func TriggerGenerate(ctx context.Context, clients *shared.ClientFactory, app types.App) (*types.DeployedTrigger, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "cmd.TriggerGenerate")
	defer span.Finish()

	// FIXME: Stop relying on context getters and setters
	token := config.GetContextToken(ctx)
	args := api.TriggerListRequest{
		AppID: app.AppID,
		Limit: 4, // Limit to improve performance for apps with many triggers
	}
	existingTriggers, _, err := clients.API().WorkflowsTriggersList(ctx, token, args)
	if err != nil {
		return nil, err
	}

	err = outputTriggersList(ctx, existingTriggers, nil, clients, app, args.Cursor, args.Type)
	if err != nil {
		return nil, err
	}

	if len(existingTriggers) > 0 {
		return nil, nil
	}

	triggerPaths := getTriggerPaths(&clients.SDKConfig)

	_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "Create a trigger",
		Secondary: []string{
			fmt.Sprintf("Searching for trigger definition files under '%s'...", strings.Join(triggerPaths, ", ")),
		},
	})))

	triggerFilePaths, err := getFullyQualifiedTriggerFilePaths(ctx, clients, triggerPaths)
	if err != nil {
		return nil, err
	}

	if len(triggerFilePaths) == 0 { // Skip prompt if there are no trigger files
		return nil, nil
	}

	const selectOptionSkip = "Do not create a trigger"
	selectOptions := append(triggerFilePaths, selectOptionSkip)

	var selectedTriggerDef string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose a trigger definition file:", selectOptions, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("trigger-def"),
		Required: true,
	})
	if err != nil {
		return nil, err
	} else if selection.Flag {
		selectedTriggerDef = selection.Option
	} else if selection.Prompt {
		if selection.Option == selectOptionSkip {
			return nil, nil
		}
		selectedTriggerDef = selection.Option
	}

	triggerArg, err := triggerRequestFromDef(ctx, clients, createCmdFlags{triggerDef: selectedTriggerDef}, app.IsDev)
	if err != nil {
		return nil, err
	}

	// Fix the app ID selected from the menu. In the --trigger-def case, this lets you use the same
	// def file for dev and prod.
	triggerArg.WorkflowAppID = app.AppID

	createdTrigger, err := clients.API().WorkflowsTriggersCreate(ctx, token, triggerArg)
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "Trigger successfully created!",
	}))
	trigs, err := sprintTrigger(ctx, createdTrigger, clients, true, app)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", strings.Join(trigs, "\n"))
	fmt.Println()

	return &createdTrigger, nil
}

// ShowTriggers returns if information for TriggerGenerate should be shown
func ShowTriggers(clients *shared.ClientFactory, hideTriggersFlag bool) bool {
	return !hideTriggersFlag &&
		clients.IO.IsTTY() &&
		clients.SDKConfig.Hooks.GetTrigger.IsAvailable()
}

// getTriggerPaths returns an array of file paths to search for
// trigger definition files. The file paths can support the Golang
// Glob file pattern and can be customized in the hooks.json file.
func getTriggerPaths(sdkCLIConfig *hooks.SDKCLIConfig) []string {
	if sdkCLIConfig == nil {
		return defaultTriggerPaths
	}

	// List of trigger file patterns to return
	var triggerPaths []string

	// Get and sanitize the file patterns
	configTriggerPaths := sdkCLIConfig.Config.TriggerPaths
	for _, configTriggerPath := range configTriggerPaths {
		pattern := strings.TrimSpace(configTriggerPath)
		if len(pattern) == 0 {
			continue
		}
		triggerPaths = append(triggerPaths, pattern)
	}

	if len(triggerPaths) <= 0 {
		return defaultTriggerPaths
	}

	return triggerPaths
}

// getFullyQualifiedTriggerFilePaths returns an array of file paths that
// are validated
func getFullyQualifiedTriggerFilePaths(ctx context.Context, clients *shared.ClientFactory, triggerPaths []string) ([]string, error) {
	projectDir, err := clients.Os.Getwd()
	if err != nil {
		return nil, err
	}

	var triggerFilePaths = []string{}

	for _, triggerPath := range triggerPaths {
		// get an array of filepaths matching the glob trigger path pattern, ignoring errors
		files, _ := clients.Os.Glob(filepath.Join(projectDir, triggerPath))

		// Accept valid paths, ignore invalid ones and log
		for _, file := range files {
			if isValidTriggerFilePath(file) {
				triggerFilePaths = append(triggerFilePaths, file)
			} else {
				clients.IO.PrintDebug(ctx, fmt.Sprintf("Ignoring invalid trigger file path: %v", file))
			}
		}
	}

	if len(triggerFilePaths) <= 0 {
		clients.IO.PrintInfo(ctx, false, style.SectionSecondaryf(
			"No trigger definition files found\nLearn more about triggers:\nhttps://tools.slack.dev/deno-slack-sdk/guides/creating-link-triggers",
		))
		return nil, nil
	} else {
		clients.IO.PrintInfo(ctx, false, style.SectionSecondaryf(
			"Found %d trigger definition %s", len(triggerFilePaths), style.Pluralize("file", "files", len(triggerFilePaths)),
		))
	}

	relativeFilePaths, err := getRelativeFilePaths(triggerFilePaths, projectDir)
	if err != nil {
		return nil, err
	}

	return relativeFilePaths, nil
}

// getRelativeFilePaths updates a list of trigger file paths to relative to the supplied project directory
func getRelativeFilePaths(triggerFilePaths []string, projectDir string) ([]string, error) {
	relativeTriggerFilePaths := make([]string, len(triggerFilePaths))

	for i, triggerFilePath := range triggerFilePaths {
		if relativeFilePath, err := filepath.Rel(projectDir, triggerFilePath); err == nil {
			relativeTriggerFilePaths[i] = relativeFilePath
		} else {
			return nil, err
		}
	}
	return relativeTriggerFilePaths, nil
}

// isValidTriggerFilePath returns true if no defaultTriggerFilePathMatchers match triggerPath supplied
func isValidTriggerFilePath(triggerPath string) bool {

	for _, invalidPathMatcher := range defaultTriggerFilePathMatchers {
		r := regexp.MustCompile(invalidPathMatcher)
		if r.MatchString(triggerPath) {
			return false
		}
	}
	return true
}

func validateCreateCmdFlags(ctx context.Context, clients *shared.ClientFactory, createFlags *createCmdFlags) error {
	if createFlags.triggerDef != "" {
		exists, err := afero.Exists(clients.Fs, createFlags.triggerDef)
		if err != nil {
			return err
		}
		if !exists {
			return slackerror.New(slackerror.ErrTriggerNotFound).WithMessage("File not found: %s", createFlags.triggerDef)
		}
		var details []slackerror.ErrorDetail
		var mismatchedFlagDetail = func(flag string) slackerror.ErrorDetail {
			return slackerror.ErrorDetail{
				Message: fmt.Sprintf("A value for the --%s flag was included in the command", flag),
			}
		}
		if clients.Config.Flags.Lookup("description").Changed {
			details = append(details, mismatchedFlagDetail("description"))
		}
		if clients.Config.Flags.Lookup("interactivity").Changed {
			details = append(details, mismatchedFlagDetail("interactivity"))
		}
		if clients.Config.Flags.Lookup("interactivity-name").Changed {
			details = append(details, mismatchedFlagDetail("interactivity-name"))
		}
		if clients.Config.Flags.Lookup("title").Changed {
			details = append(details, mismatchedFlagDetail("title"))
		}
		if clients.Config.Flags.Lookup("workflow").Changed {
			details = append(details, mismatchedFlagDetail("workflow"))
		}
		if len(details) > 0 {
			details = append([]slackerror.ErrorDetail{{
				Message: "The --trigger-def flag overrides other property setting flags",
			}}, details...)
			return slackerror.New(slackerror.ErrMismatchedFlags).
				WithDetails(details).
				WithRemediation("Use either the --trigger-def flag or the property setting flags (e.g. --title)")
		}
	}

	if createFlags.triggerDef == "" && createFlags.workflow == "" {
		return maybeSetTriggerDefFlag(ctx, clients, createFlags)
	}

	if createFlags.description == "" {
		createFlags.description = fmt.Sprintf("Runs the '%s' workflow", createFlags.workflow)
	}

	return nil
}

func maybeSetTriggerDefFlag(ctx context.Context, clients *shared.ClientFactory, createFlags *createCmdFlags) error {
	triggerPaths := getTriggerPaths(&clients.SDKConfig)

	fmt.Printf("%s", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  fmt.Sprintf("Searching for trigger definition files under '%s'...", strings.Join(triggerPaths, ", ")),
	}))

	triggerFilePaths, err := getFullyQualifiedTriggerFilePaths(ctx, clients, triggerPaths)
	if err != nil {
		return err
	}

	if len(triggerFilePaths) == 0 {
		return slackerror.New(slackerror.ErrMismatchedFlags).WithMessage("--workflow or --trigger-def is required")
	}

	selection, err := clients.IO.SelectPrompt(ctx, "Choose a trigger definition file:", triggerFilePaths, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("trigger-def"),
		Required: true,
	})
	if err != nil {
		return err
	} else {
		createFlags.triggerDef = selection.Option
	}

	return nil
}
