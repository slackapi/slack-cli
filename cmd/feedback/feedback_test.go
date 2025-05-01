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
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbackCommand(t *testing.T) {
	t.Run("when there is only one survey option", func(t *testing.T) {
		surveys := map[string]SlackSurvey{
			SlackPlatformFeedback: {
				Name:              SlackPlatformFeedback,
				PromptDisplayText: "Please complete this survey",
				URL:               url.URL{RawPath: "https://survey.com"},
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
			},
		}

		// Prepare mocks
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()

		pcm := &config.ProjectConfigMock{}
		pcm.On("GetProjectID", mock.Anything).Return("projectID", nil)
		pcm.On("GetSurveyConfig", mock.Anything, SlackPlatformFeedback).Return(config.SurveyConfig{}, nil)
		clientsMock.Config.ProjectConfig = pcm

		scm := &config.SystemConfigMock{}
		scm.On("GetSurveyConfig", mock.Anything, SlackPlatformFeedback).Return(config.SurveyConfig{}, nil)
		scm.On("GetSystemID", mock.Anything).Return("systemID", nil)
		scm.On("SetSurveyConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		clientsMock.Config.SystemConfig = scm

		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Open in browser?", mock.Anything).Return(true)
		clientsMock.Browser.On("OpenURL", "https://survey.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli").Return(nil)

		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		_surveys := SurveyStore
		SurveyStore = surveys
		defer func() { SurveyStore = _surveys }()

		// Execute test
		cmd := NewFeedbackCommand(clients)
		err := runFeedbackCommand(ctx, clients, cmd)

		// Assertions
		assert.NoError(t, err)
		clientsMock.Browser.AssertCalled(t, "OpenURL", "https://survey.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli")
		clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.FeedbackMessage, []string{SlackPlatformFeedback})
	})

	t.Run("when there are multiple survey options", func(t *testing.T) {
		surveys := map[string]SlackSurvey{
			"A_test": {
				Name:              "A_test",
				PromptDisplayText: "A",
				URL:               url.URL{RawPath: "https://A.com"},
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
			},
			SlackPlatformFeedback: {
				Name:              SlackPlatformFeedback,
				PromptDisplayText: "PlatformSurvey survey",
				URL:               url.URL{RawPath: "https://survey.com"},
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
			},
			"B_test": {
				Name:              "B_test",
				PromptDisplayText: "B",
				URL:               url.URL{RawPath: "https://B.com"},
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.ProjectConfig
				},
			},
		}

		// Prepare mocks
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()

		scm := &config.SystemConfigMock{}
		scm.On("GetSurveyConfig", mock.Anything, mock.Anything).Return(config.SurveyConfig{}, nil)
		scm.On("GetSystemID", mock.Anything).Return("systemID", nil)
		scm.On("SetSurveyConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		clientsMock.Config.SystemConfig = scm

		pcm := &config.ProjectConfigMock{}
		pcm.On("GetSurveyConfig", mock.Anything, mock.Anything).Return(config.SurveyConfig{}, nil)
		pcm.On("GetProjectID", mock.Anything).Return("projectID", nil)
		clientsMock.Config.ProjectConfig = pcm

		clientsMock.IO.On("SelectPrompt", mock.Anything, "What type of feedback would you like to give?\n", []string{"A", "B", "PlatformSurvey survey"}, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("name"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Index:  2,
		}, nil)

		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Open in browser?", mock.Anything).Return(true)
		clientsMock.Browser.On("OpenURL", "https://survey.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli").Return(nil)

		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		_surveys := SurveyStore
		SurveyStore = surveys
		defer func() { SurveyStore = _surveys }()

		// Execute test
		cmd := NewFeedbackCommand(clients)
		err := runFeedbackCommand(ctx, clients, cmd)
		assert.NoError(t, err)

		// Assertions
		clientsMock.Browser.AssertCalled(t, "OpenURL", "https://survey.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli")
		clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.FeedbackMessage, []string{SlackPlatformFeedback})
	})
}

func TestShowSurveyMessages(t *testing.T) {
	t.Run("surveys asked or not asked based on the stored config", func(t *testing.T) {
		surveys := map[string]SlackSurvey{
			// Should be asked once; already asked
			"A": {
				Name:      "A",
				URL:       url.URL{RawPath: "https://A.com"},
				Frequency: Once,
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
				Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
					msg := fmt.Sprintf("%s Would you like to take a minute to tell us about A?", style.Emoji("love_letter"))
					return clients.IO.ConfirmPrompt(ctx, msg, false)
				},
			},
			// Should be asked once; not yet asked
			"B": {
				Name:      "B",
				URL:       url.URL{RawPath: "https://B.com"},
				Frequency: Once,
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.ProjectConfig
				},
				Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
					return clients.IO.ConfirmPrompt(ctx, "Would you like to take a minute to tell us about B?", false)
				},
			},
			// Ask monthly; it's been over a month
			"C": {
				Name:      "C",
				URL:       url.URL{RawPath: "https://C.com"},
				Frequency: Monthly,
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
				Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
					return clients.IO.ConfirmPrompt(ctx, "Would you like to take a minute to tell us about C?", false)
				},
			},
			// Ask monthly; it's only been a day since last asked
			"D": {
				Name:      "D",
				URL:       url.URL{RawPath: "https://D.com"},
				Frequency: Monthly,
				Config: func(clients *shared.ClientFactory) SurveyConfigInterface {
					return clients.Config.SystemConfig
				},
				Ask: func(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
					return clients.IO.ConfirmPrompt(ctx, "Would you like to take a minute to tell us about D?", false)
				},
			},
		}

		// Prepare mocks
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()

		scm := &config.SystemConfigMock{}
		pcm := &config.ProjectConfigMock{}

		oneMonthAgo := time.Now().Unix() - (60 * 60 * 24 * 32)
		oneMonthAgoTimestamp := time.Unix(oneMonthAgo, 0).Format(time.RFC3339)

		// A
		scm.On("GetSurveyConfig", mock.Anything, "A").Return(config.SurveyConfig{AskedAt: oneMonthAgoTimestamp}, nil).Once()

		// B
		pcm.On("GetSurveyConfig", mock.Anything, "B").Return(config.SurveyConfig{}, slackerror.New(slackerror.ErrSurveyConfigNotFound)).Once()
		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Would you like to take a minute to tell us about B?", mock.Anything).Return(true)
		scm.On("GetSystemID", mock.Anything).Return("systemID", nil).Once()
		pcm.On("GetProjectID", mock.Anything).Return("projectID", nil).Once()
		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Open in browser?", mock.Anything).Return(true).Once()
		clientsMock.Browser.On("OpenURL", "https://B.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli").Return(nil).Once()
		pcm.On("SetSurveyConfig", mock.Anything, "B", mock.Anything).Return(nil).Once()

		// C
		scm.On("GetSurveyConfig", mock.Anything, "C").Return(config.SurveyConfig{AskedAt: oneMonthAgoTimestamp}, nil).Once()
		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Would you like to take a minute to tell us about C?", mock.Anything).Return(true)
		scm.On("GetSystemID", mock.Anything).Return("systemID", nil).Once()
		pcm.On("GetProjectID", mock.Anything).Return("projectID", nil).Once()
		clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Open in browser?", mock.Anything).Return(true).Once()
		clientsMock.Browser.On("OpenURL", "https://C.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli").Return(nil).Once()
		scm.On("SetSurveyConfig", mock.Anything, "C", mock.Anything).Return(nil).Once()

		// D
		oneDayAgo := time.Now().Unix() - (60 * 60 * 24)
		oneDayAgoTimestamp := time.Unix(oneDayAgo, 0).Format(time.RFC3339)
		scm.On("GetSurveyConfig", mock.Anything, "D").Return(config.SurveyConfig{AskedAt: oneDayAgoTimestamp}, nil).Once()

		clientsMock.Config.SystemConfig = scm
		clientsMock.Config.ProjectConfig = pcm

		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		_surveys := SurveyStore
		SurveyStore = surveys
		defer func() { SurveyStore = _surveys }()

		// Execute test
		err := ShowSurveyMessages(ctx, clients)
		assert.NoError(t, err)
		clientsMock.Browser.AssertCalled(t, "OpenURL", "https://B.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli")
		clientsMock.Browser.AssertCalled(t, "OpenURL", "https://C.com?project_id=projectID&system_id=systemID&utm_medium=cli&utm_source=cli")
	})
}

func Test_Feedback_FeedbackCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		// DEPRECATED(semver:major): Support the deprecated survey name for backwards compatibility
		"supports deprecated --name platform-improvements": {
			CmdArgs: []string{"--name", "platform-improvements"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupFeedbackCommandMocks(t, ctx, cm, cf)
			},
			ExpectedOutputs: []string{
				"feedback@slack.com",
				"https://docs.slack.dev/developer-support",
			},
		},
		"supports --name slack-cli": {
			CmdArgs: []string{"--name", "slack-cli"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupFeedbackCommandMocks(t, ctx, cm, cf)
			},
			ExpectedOutputs: []string{
				"Ask questions, submit issues, or suggest features for the Slack CLI",
				"https://github.com/slackapi/slack-cli/issues",
			},
		},
		"supports --name slack-platform": {
			CmdArgs: []string{"--name", "slack-platform"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupFeedbackCommandMocks(t, ctx, cm, cf)
			},
			ExpectedOutputs: []string{
				"feedback@slack.com",
				"https://docs.slack.dev/developer-support",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewFeedbackCommand(cf)
		return cmd
	})
}

// setupFeedbackCommandMocks prepares common mocks for these tests
func setupFeedbackCommandMocks(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cm.AddDefaultMocks()

	scm := &config.SystemConfigMock{}
	scm.On("GetSurveyConfig", mock.Anything, mock.Anything).Return(config.SurveyConfig{}, nil)
	scm.On("GetSystemID", mock.Anything).Return("systemID", nil)
	scm.On("SetSurveyConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cm.Config.SystemConfig = scm

	pcm := &config.ProjectConfigMock{}
	pcm.On("GetSurveyConfig", mock.Anything, mock.Anything).Return(config.SurveyConfig{}, nil)
	pcm.On("GetProjectID", mock.Anything).Return("projectID", nil)
	cm.Config.ProjectConfig = pcm

	cm.IO.On("ConfirmPrompt", mock.Anything, "Open in browser?", mock.Anything).Return(false)
}
