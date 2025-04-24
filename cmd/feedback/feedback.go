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

package feedback

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var surveyNameFlag string
var noPromptFlag bool

type SlackSurvey struct {
	// Name is the survey identifier
	Name string
	// PromptDisplayText is displayed as the `feedback` command prompt option
	PromptDisplayText string
	// PromptDescription is displayed beneath the `feedback` command prompt option
	PromptDescription string
	// SkipQueryParams is a flag to skip adding query params to the survey URL
	SkipQueryParams bool
	// URL is the survey URL
	URL url.URL
	// Config returns either the project-level or system-level survey config
	Config func(clients *shared.ClientFactory) SurveyConfigInterface
	// Frequency is how often we should ask users to complete the survey
	Frequency Frequency
	// Info prints additional information about the survey; displayed when the option is selected in `feedback`
	// Info is optional
	Info func(ctx context.Context, clients *shared.ClientFactory)
	// Trace is a string consumed by tests to confirm that the Info text was displayed
	Trace string
	// Ask either prints text or prompts the user to complete the survey
	// Potentially displayed after `run`/`deploy`/`doctor` (or other places where ShowSurveyMessages is called)
	Ask func(ctx context.Context, clients *shared.ClientFactory) (bool, error)
}

// Frequency defines how often we want to ask the user to complete the survey
type Frequency int

const (
	Always  Frequency = iota // We always want to ask
	Once                     // Ask user to complete the survey only once
	Monthly                  // Ask user to complete the survey once a month
	Never                    // Do not ask the user to complete the survey
)

// Supported survey names
const (
	SlackCLIFeedback      = "slack-cli-feedback"
	SlackPlatformFeedback = "platform-improvements"
)

type SurveyConfigInterface interface {
	GetSurveyConfig(ctx context.Context, name string) (config.SurveyConfig, error)
	SetSurveyConfig(ctx context.Context, name string, surveyConfig config.SurveyConfig) error
}

// SurveyStore stores all available surveys.
// New surveys should be added here.
var SurveyStore = map[string]SlackSurvey{
	// SlackCLIFeedback asks for Slack CLI feedback using GitHub Issues
	SlackCLIFeedback: {
		Name:              SlackCLIFeedback,
		PromptDisplayText: "Slack CLI",
		PromptDescription: "Questions, issues, and feature requests about the Slack CLI",
		SkipQueryParams:   true,
		URL: url.URL{
			RawPath: "https://github.com/slackapi/slack-cli/issues",
		},
		Info: func(ctx context.Context, clients *shared.ClientFactory) {
			clients.IO.PrintInfo(ctx, false, fmt.Sprintf(
				"%s\n%s\n",
				style.Secondary("Ask questions, submit issues, or suggest features for the SLack CLI:"),
				style.Secondary(style.Highlight("https://github.com/slackapi/slack-cli/issues")),
			))
		},
		Trace: slacktrace.FeedbackMessage,
		Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
			clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
				Emoji: "love_letter",
				Text:  "We would love to know how things are going",
				Secondary: []string{
					"Share your experience with " + style.Commandf("feedback --name slack-cli-feedback", false),
				},
			}))
			return false, nil
		},
		Frequency: Never,
		Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
			return clients.Config.SystemConfig
		},
	},
	// SlackPlatformFeedback asks for general developer experience feedback
	SlackPlatformFeedback: {
		Name:              SlackPlatformFeedback,
		PromptDisplayText: "Slack Platform",
		PromptDescription: "Developer support for the Slack Platform, Slack API, Block Kit, and more",
		URL:               url.URL{RawPath: "https://docs.slack.dev/developer-support"},
		Info: func(ctx context.Context, clients *shared.ClientFactory) {
			clients.IO.PrintInfo(ctx, false, fmt.Sprintf(
				"%s\n%s\n",
				style.Secondary("You can send us a message at "+style.Highlight(email)),
				style.Secondary("Or, share your experiences at "+style.Highlight("https://docs.slack.dev/developer-support")),
			))
		},
		Trace: slacktrace.FeedbackMessage,
		Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
			clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
				Emoji: "love_letter",
				Text:  "We would love to know how things are going",
				Secondary: []string{
					"Share your development experience with " + style.Commandf("feedback", false),
				},
			}))
			return false, nil
		},
		Frequency: Always,
		Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
			return clients.Config.SystemConfig
		},
	},
}

// SetAskedAtTimestamp writes a timestamp for when the survey was last asked to the project or system level config
func (s SlackSurvey) SetAskedAtTimestamp(ctx context.Context, clients *shared.ClientFactory) error {
	t := time.Now()
	err := s.Config(clients).SetSurveyConfig(ctx, s.Name, config.SurveyConfig{AskedAt: t.Format(time.RFC3339)})
	if err != nil {
		return err
	}
	return nil
}

// SetCompletedAtTimestamp writes a timestamp for when the survey was asked and completed to the project or system level config
func (s SlackSurvey) SetCompletedAtTimestamp(ctx context.Context, clients *shared.ClientFactory, name string) error {
	t := time.Now()
	timestamp := t.Format(time.RFC3339)
	err := s.Config(clients).SetSurveyConfig(ctx, name, config.SurveyConfig{AskedAt: timestamp, CompletedAt: timestamp})
	if err != nil {
		return err
	}
	return nil
}

// ShouldAsk returns true if we should ask the user the complete the survey
func (s SlackSurvey) ShouldAsk(cfg config.SurveyConfig) (bool, error) {
	if s.Frequency == Never {
		return false, nil
	}

	if cfg.AskedAt == "" { // survey has never been asked before
		return true, nil
	}

	if s.Frequency == Always {
		return true, nil
	}

	if s.Frequency == Once {
		return false, nil // we've already asked
	}

	t, err := time.Parse(time.RFC3339, cfg.AskedAt)
	if err != nil {
		return false, err
	}
	if s.Frequency == Monthly {
		secondsInAMonth := int64(60 * 60 * 24 * 31)
		secondsElapsedSinceAsked := time.Now().Unix() - t.Unix()
		return secondsElapsedSinceAsked > secondsInAMonth, nil // Return true if we asked over a month ago
	}

	return false, nil
}

const (
	email       = "feedback@slack.com"
	leadMessage = "Thanks for taking a moment to share your feedback!"
)

func NewFeedbackCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "feedback",
		Aliases: []string{},
		Short:   "Share feedback about your experience or project",
		Long:    "Help us make the Slack Platform better with your feedback",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "feedback", Meaning: "Choose to give feedback on part of the Slack Platform"},
			{Command: "feedback --name slack-cli-feedback", Meaning: "Give feedback on the Slack CLI"},
		}),
		PreRun: func(cmd *cobra.Command, args []string) {
			clients.Config.SetFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runFeedbackCommand(ctx, clients, cmd)
		},
	}

	// Initialize flags

	surveyNames := []string{}
	for _, s := range SurveyStore {
		surveyNames = append(surveyNames, s.Name)
	}
	sort.Strings(surveyNames)
	nameFlagDescription := style.Sectionf(style.TextSection{
		Text:      "name of the feedback:",
		Secondary: surveyNames,
	})
	cmd.Flags().StringVar(&surveyNameFlag, "name", "", nameFlagDescription)

	cmd.Flags().BoolVar(&noPromptFlag, "no-prompt", false, "run command without prompts")

	return cmd
}

// runFeedbackCommand will open the user's browser to the feedback survey webpage.
func runFeedbackCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	if len(SurveyStore) == 0 {
		clients.IO.PrintInfo(ctx, false, "No feedback options currently available; please try again later")
		return nil
	}

	surveyNames, surveyPromptOptions := initSurveyOpts(ctx, clients, SurveyStore)

	if _, ok := SurveyStore[surveyNameFlag]; !ok && surveyNameFlag != "" {
		return slackerror.New("invalid_survey_name").
			WithMessage("Invalid feedback name provided: %s", surveyNameFlag).
			WithRemediation("View the feedback options with %s", style.Commandf("feedback --help", false))
	}

	if surveyNameFlag == "" && noPromptFlag {
		return slackerror.New("survey_name_required").
			WithMessage("Please provide a feedback name or remove the --no-prompt flag").
			WithRemediation("View feedback options with %s", style.Commandf("feedback --help", false))
	}

	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "love_letter",
		Text:  leadMessage,
	}))

	var err error

	surveyName := surveyNameFlag
	if surveyName == "" {
		if len(surveyNames) == 1 {
			surveyName = surveyNames[0]
		} else {
			surveyName, err = chooseSurveyPrompt(ctx, clients, surveyNames, surveyPromptOptions)
			if err != nil {
				return err
			}
		}
	}

	err = executeSurvey(ctx, clients, SurveyStore[surveyName])
	if err != nil {
		return err
	}

	return nil
}

// initSurveyOpts prepares prompt options based on the survey store
func initSurveyOpts(ctx context.Context, clients *shared.ClientFactory, surveys map[string]SlackSurvey) ([]string, []string) {
	var sortedSurveys []SlackSurvey

	// Sort survey options consistently
	for _, s := range SurveyStore {
		sortedSurveys = append(sortedSurveys, s)
	}
	sort.Slice(sortedSurveys, func(i, j int) bool {
		return sortedSurveys[i].PromptDisplayText < sortedSurveys[j].PromptDisplayText
	})

	// Initialize survey options
	var names []string
	var opts []string
	for _, s := range sortedSurveys {
		if s.Config == nil {
			clients.IO.PrintDebug(ctx, fmt.Sprintf("survey config not set; skipping %s", s.Name))
			continue
		}
		names = append(names, s.Name)
		cfg, err := s.Config(clients).GetSurveyConfig(ctx, s.Name)
		if err != nil {
			if !slackerror.IsErrorType(err, slackerror.ErrSurveyConfigNotFound) {
				clients.IO.PrintDebug(ctx, "Error getting survey config for %s: %s", s.Name, err)
			}
			opts = append(opts, s.PromptDisplayText)
			continue
		}
		if cfg.CompletedAt == "" {
			opts = append(opts, s.PromptDisplayText)
			continue
		}
		t, err := time.Parse(time.RFC3339, cfg.CompletedAt)
		if err != nil {
			clients.IO.PrintDebug(ctx, err.Error())
			opts = append(opts, s.PromptDisplayText)
			continue
		}
		opts = append(opts, fmt.Sprintf("%s (completed %s)", s.PromptDisplayText, t.Format("January 2, 2006")))
	}

	return names, opts
}

// executeSurvey prints a message, opens the survey URL and marks the survey as completed
func executeSurvey(ctx context.Context, clients *shared.ClientFactory, s SlackSurvey) error {
	// Display survey info
	if s.Info != nil {
		s.Info(ctx, clients)
	}
	clients.IO.PrintTrace(ctx, s.Trace, s.Name)

	var err error
	var ok bool
	if !noPromptFlag {
		ok, err = clients.IO.ConfirmPrompt(ctx, "Open in browser?", true)
		if err != nil {
			return err
		}
	}

	url := s.URL.RawPath

	if !s.SkipQueryParams {
		url, err = addQueryParams(ctx, clients, s.URL.RawPath)
		if err != nil {
			return err
		}
	}

	if ok { // Open survey in browser
		clients.Browser().OpenURL(url)
	} else { // Print survey URL
		clients.IO.PrintInfo(ctx, false, fmt.Sprint("Feedback URL: \n", style.Secondary(url)))
	}

	// Record completion
	return s.SetCompletedAtTimestamp(ctx, clients, s.Name)
}

// chooseSurveyPrompt prompts the user to select a survey
func chooseSurveyPrompt(ctx context.Context, clients *shared.ClientFactory, surveyNames []string, surveyPromptOptions []string) (string, error) {
	msg := "What type of feedback would you like to give?\n"

	var survey string
	selection, err := clients.IO.SelectPrompt(ctx, msg, surveyPromptOptions,
		iostreams.SelectPromptConfig{
			Flag:     clients.Config.Flags.Lookup("name"),
			PageSize: 4,
			Description: func(value string, index int) string {
				if index < len(surveyNames) {
					return SurveyStore[surveyNames[index]].PromptDescription
				}
				return ""
			}})
	if err != nil {
		return "", err
	} else if selection.Flag {
		survey = selection.Option
	} else if selection.Prompt {
		survey = surveyNames[selection.Index]
	}

	fmt.Println()

	return survey, nil
}

// addQueryParams adds common query params to the survey URL for tracking
func addQueryParams(ctx context.Context, clients *shared.ClientFactory, originalURL string) (string, error) {
	u, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("utm_medium", "cli")
	q.Set("utm_source", "cli")

	systemID, err := clients.Config.SystemConfig.GetSystemID(ctx)
	if err != nil {
		return "", err
	}
	q.Set("system_id", systemID)

	projectID, err := clients.Config.ProjectConfig.GetProjectID(ctx)
	if err == nil {
		q.Set("project_id", projectID)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ShowSurveyMessages displays a message or prompt for feedback for one or more surveys
func ShowSurveyMessages(ctx context.Context, clients *shared.ClientFactory) error {
	for _, s := range SurveyStore {
		cfg, err := s.Config(clients).GetSurveyConfig(ctx, s.Name)
		if err != nil {
			if !slackerror.IsErrorType(err, slackerror.ErrSurveyConfigNotFound) {
				clients.IO.PrintDebug(ctx, "Error getting survey config for %s: %s", s.Name, err)
				continue
			}
		}
		ok, err := s.ShouldAsk(cfg)
		if err != nil {
			clients.IO.PrintDebug(ctx, "Error checking survey config for %s: %s", s.Name, err)
			continue
		}
		if !ok {
			continue
		}
		shouldExecuteSurvey, err := s.Ask(ctx, clients)
		if err != nil {
			return err
		}
		if !shouldExecuteSurvey {
			err = s.SetAskedAtTimestamp(ctx, clients)
			if err != nil {
				return err
			}
			continue
		}
		err = executeSurvey(ctx, clients, s)
		if err != nil {
			return err
		}
	}
	return nil
}

// ShowFeedbackMessageOnTerminate prints a message asking for user feedback when
// an interrupt signal is received, flushing the ^C ctrl+C character in the process.
func ShowFeedbackMessageOnTerminate(ctx context.Context, clients *shared.ClientFactory) {
	err := ShowSurveyMessages(ctx, clients)
	if err != nil {
		clients.IO.PrintError(ctx, err.Error())
	}
}
