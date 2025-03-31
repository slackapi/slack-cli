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

package platform

import (
	"fmt"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/platform"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var tailArg bool
var pollingIntervalS int
var idleTimeoutM int
var browser bool

// Filters
var limit int
var minDateCreated int64
var maxDateCreated int64
var minLevel string
var eventType string
var componentType string
var componentId string
var source string
var traceId string

// Create handle to the function for testing
// TODO - Stopgap until we learn the correct way to structure our code for testing.
var activityFunc = platform.Activity
var runActivityCommandFunc = runActivityCommand
var appSelectPromptFunc = prompts.AppSelectPrompt

func NewActivityCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activity",
		Aliases: []string{"log", "logs"},
		Short:   "Display the app activity logs from the Slack Platform",
		Long:    `Display the app activity logs from the Slack Platform`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "platform activity", Meaning: "Display app activity logs for an app"},
			{Command: "platform activity -t", Meaning: "Continuously poll for new activity logs"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivityCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().BoolVarP(&tailArg, "tail", "t", false, "continuously poll for new activity")
	cmd.Flags().IntVarP(&pollingIntervalS, "interval", "i", platform.ACTIVITY_POLLING_INTERVAL_SECONDS, "polling interval in seconds")
	cmd.Flags().IntVarP(&idleTimeoutM, "idle", "", platform.ACTIVITY_IDLE_TIMEOUT, "time to poll without results before exiting\n  in minutes")
	cmd.Flags().BoolVarP(&browser, "browser", "", false, "open the default web browser to the log activity")
	cmd.Flags().StringVar(&minLevel, "level", "", fmt.Sprintf("minimum log level to display (default \"%s\")\n  (trace, debug, info, warn, error, fatal)", platform.ACTIVITY_MIN_LEVEL))
	cmd.Flags().IntVar(&limit, "limit", platform.ACTIVITY_LIMIT, "limit the amount of logs retrieved")
	cmd.Flags().Int64Var(&minDateCreated, "min-date-created", 0, "minimum timestamp to filter\n  (unix timestamp in microseconds)")
	cmd.Flags().Int64Var(&maxDateCreated, "max-date-created", 0, "maximum timestamp to filter\n  (unix timestamp in microseconds)")
	cmd.Flags().StringVar(&eventType, "event", "", "event type to filter")
	cmd.Flags().StringVar(&componentType, "component", "", "component type to filter")
	cmd.Flags().StringVar(&componentId, "component-id", "", "component id to filter\n  (either a function id or workflow id)")
	cmd.Flags().StringVar(&source, "source", "", "source (slack or developer) to filter")
	cmd.Flags().StringVar(&traceId, "trace-id", "", "trace id to filter")

	// Hidden flags
	_ = cmd.Flags().MarkHidden("browser") // Hide until editing apps from the App Config UI is available

	// Set defaults
	if minLevel == "" {
		minLevel = platform.ACTIVITY_MIN_LEVEL
	}

	return cmd
}

// runActivityCommand executes the platform activity command
func runActivityCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	var ctx = cmd.Context()

	if maxDateCreated != 0 && tailArg {
		return slackerror.New(slackerror.ErrMismatchedFlags).WithMessage("--tail can not be used with --max-date-created")
	}

	// Prompt for an installed app
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	ctx = config.SetContextToken(ctx, selection.Auth.Token)

	activityArgs := types.ActivityArgs{
		TeamId:            selection.Auth.TeamID,
		AppId:             selection.App.AppID,
		TailArg:           tailArg,
		Browser:           browser,
		PollingIntervalMS: pollingIntervalS * 1000,
		IdleTimeoutM:      idleTimeoutM,
		Limit:             limit,
		MinDateCreated:    minDateCreated,
		MaxDateCreated:    maxDateCreated,
		MinLevel:          minLevel,
		EventType:         eventType,
		ComponentType:     componentType,
		ComponentId:       componentId,
		Source:            source,
		TraceId:           traceId,
	}

	log := newActivityLogger(cmd)
	if err := activityFunc(ctx, clients, log, activityArgs); err != nil {
		return err
	}

	return nil
}

// newActivityLogger creates a logger instance to receive event notifications
func newActivityLogger(cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			default:
				// Ignore the event
			}
		},
	)
}
