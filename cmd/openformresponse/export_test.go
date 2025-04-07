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

package openformresponse

import (
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var appId = "appId"
var token = "token"
var installedProdApp = prompts.SelectedApp{Auth: types.SlackAuth{Token: token}, App: types.App{AppID: appId}}

func TestExportCommand(t *testing.T) {
	var appSelectTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"missing --workflow": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedErrorStrings: []string{"--workflow", "required"},
		},
		"with --step-id, API succeeds": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow", "--step-id", "stepId"},
			ExpectedOutputs: []string{"Slackbot will DM you with a CSV file once it's ready"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				cm.ApiInterface.On("StepsResponsesExport", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "StepsResponsesExport", mock.Anything, token, "#/workflows/my_workflow", appId, "stepId")
			},
		},
		"with --step-id, API fails": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow", "--step-id", "stepId"},
			ExpectedErrorStrings: []string{"failed"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				cm.ApiInterface.On("StepsResponsesExport", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed"))
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "StepsResponsesExport", mock.Anything, token, "#/workflows/my_workflow", appId, "stepId")
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewExportCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func setupMockCreateAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = exportAppSelectPromptFunc
	exportAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		exportAppSelectPromptFunc = originalPromptFunc
	}
}
