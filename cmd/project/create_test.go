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

package project

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CreateClientMock struct {
	mock.Mock
}

func (m *CreateClientMock) Create(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, createArgs create.CreateArgs) (string, error) {
	args := m.Called(ctx, clients, log, createArgs)
	return args.String(0), args.Error(1)
}

func TestCreateCommand(t *testing.T) {
	var createClientMock *CreateClientMock

	testutil.TableTestCommand(t, testutil.CommandTests{
		"creates a bolt application from prompts": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select an app:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a language:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, expected)
			},
		},
		"creates a deno application from flags": {
			CmdArgs: []string{"--template", "slack-samples/deno-starter-template"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select an app:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/deno-starter-template",
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a language:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/deno-starter-template",
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/deno-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, expected)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func TestCreateCommand_confirmExternalTemplateSelection(t *testing.T) {
	// test cases
	tests := map[string]struct {
		templateSource string
		setup          func(cm *shared.ClientsMock, scm *config.SystemConfigMock)
		expect         func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock)
	}{
		"trust, untrusted source": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(true, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "selection should be confirmed")
				}
				// Should not have prompted
				output := cm.GetCombinedOutput()
				assert.NotContains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, trusted source": {
			templateSource: "slack-samples/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should confirm the selection despite not trusting source since template source is trusted")
				}
			},
		},
		"don't trust, untrusted source, confirm": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {

				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 0, Option: "Yes"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should prompt")
				}

				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, untrusted source, do not confirm": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 2, Option: "No"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.False(t, confirmed)
				if err != nil {
					assert.Fail(t, "should have returned with no error")
				}

				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, untrusted source, confirm and don't ask again": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(nil)

				// Proceed when prompted and select don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")

				// should have confirmed
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should have returned with no error")
				}
				// Should have tried to set the source after confirmation
				scm.AssertCalled(t, "SetTrustUnknownSources", mock.Anything, true)
			},
		},
		"don't trust, untrusted source, confirm and prompt error": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(nil)

				// Proceed when prompted, don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, slackerror.New("error"))
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// does not return confirmed
				assert.False(t, confirmed)
				if err == nil {
					assert.Fail(t, "should have returned the prompt error")
				}
				// Should not have tried to set the source after confirmation
				scm.AssertNotCalled(t, "SetTrustUnknownSources", mock.Anything, true)
			},
		},
		"don't trust sources, untrusted source, confirm and config write error": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)

				// Set trust_unknown_sources should fail with an error
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(slackerror.New("an error!"))

				// Proceed when prompted, select don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// should confirm with an error
				assert.True(t, confirmed)
				if err == nil {
					assert.Fail(t, "should have returned an error")
				}
				assert.ErrorContains(t, err, "an error!")
			},
		},
	}

	// test!
	for ttName, tt := range tests {
		t.Run(ttName, func(t *testing.T) {
			// setup
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			scm := &config.SystemConfigMock{}
			tt.setup(cm, scm)
			cm.Config.SystemConfig = scm
			clients := shared.NewClientFactory(cm.MockClientFactory())
			cmd := NewCreateCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)

			// test
			template, err := create.ResolveTemplateURL(tt.templateSource)
			require.NoError(t, err)
			confirmed, err := confirmExternalTemplateSelection(cmd, clients, template)
			tt.expect(confirmed, err, cm, scm)
		})
	}
}
