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

package app

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRequestCreated is the moment a mocked request was made
var mockRequestCreated = time.Date(2026, 8, 21, 15, 4, 5, 0, time.UTC).Unix()

// mockRequestResolved is the moment a mocked request was reviewed
var mockRequestResolved = time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC).Unix()

func TestRequestCommand(t *testing.T) {
	// enableRequest turns on the experiment that gates the command
	enableRequest := func(ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
		cm.AddDefaultMocks()
		cf.SDKConfig = hooks.NewSDKConfigMock()
		cf.Config.ExperimentsFlag = []string{string(experiment.AppApprovalStatus)}
		cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
		requestAppSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, environment prompts.AppEnvironmentType, status prompts.AppInstallStatus, opts ...prompts.AppSelectOption) (prompts.SelectedApp, error) {
			return prompts.SelectedApp{
				App:  types.App{AppID: "A1234", TeamID: "T1234", TeamDomain: "teamone"},
				Auth: types.SlackAuth{Token: "xoxp-example", TeamID: "T1234", TeamDomain: "teamone"},
			}, nil
		}
	}

	restoreRequest := func() {
		requestAppSelectPromptFunc = prompts.AppSelectPrompt
		requestTeamSelectPromptFunc = prompts.PromptTeamSlackAuth
	}

	// enableRequestWithoutProject turns on the experiment outside of a project
	enableRequestWithoutProject := func(ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
		cm.AddDefaultMocks()
		cf.Config.ExperimentsFlag = []string{string(experiment.AppApprovalStatus)}
		cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
	}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors when the app-approval-status experiment is off": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cf.SDKConfig = hooks.NewSDKConfigMock()
				cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
			},
			ExpectedError: slackerror.New(slackerror.ErrExperimentRequired),
		},
		"reports a request that awaits review": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				cm.API.On("ListAppApprovalRequests", mock.Anything, "xoxp-example", "A1234", []string(nil)).
					Return(api.AppsApprovalsRequestsListResult{
						Requests: []api.AppsApprovalsRequest{
							{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
						},
					}, nil)
			},
			Teardown: restoreRequest,
			ExpectedOutputs: []string{
				"App Install Approval Requests",
				"App ID:       A1234",
				"teamone (T1234):",
				"Request ID:   Ar1234",
				"Status:       pending",
				"Requested:",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListAppApprovalRequests", mock.Anything, "xoxp-example", "A1234", []string(nil))
			},
		},
		"explains that an app was never requested": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				cm.API.On("ListAppApprovalRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(api.AppsApprovalsRequestsListResult{Requests: []api.AppsApprovalsRequest{}}, nil)
			},
			Teardown:        restoreRequest,
			ExpectedOutputs: []string{"You have not requested to install this app"},
		},
		"searches the workspaces of the provided workspace IDs": {
			CmdArgs: []string{"--workspace-ids", "T1234,T5678"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				cm.API.On("ListAppApprovalRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(api.AppsApprovalsRequestsListResult{}, nil)
			},
			Teardown: restoreRequest,
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListAppApprovalRequests", mock.Anything, "xoxp-example", "A1234", []string{"T1234", "T5678"})
			},
		},
		"returns the error of too many searched workspaces": {
			CmdArgs: []string{"--workspace-ids", strings.Join(mockRequestWorkspaceIDs(51), ",")},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				cm.API.On("ListAppApprovalRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(api.AppsApprovalsRequestsListResult{}, slackerror.New(slackerror.ErrInvalidArguments))
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrInvalidArguments),
		},
		"suggests the app flag without a project directory": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequestWithoutProject(ctx, cm, cf)
			},
			Teardown:             restoreRequest,
			ExpectedErrorStrings: []string{slackerror.ErrInvalidAppDirectory, "hooks.json", "--app A0123456789"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"checks an app named by ID outside of a project": {
			CmdArgs: []string{"--app", "A5678"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequestWithoutProject(ctx, cm, cf)
				requestTeamSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, promptText string, promptConfig *prompts.PromptTeamSlackAuthConfig) (*types.SlackAuth, error) {
					return &types.SlackAuth{Token: "xoxp-selected", TeamID: "T5678", TeamDomain: "teamtwo"}, nil
				}
				cm.API.On("ListAppApprovalRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(api.AppsApprovalsRequestsListResult{
						Requests: []api.AppsApprovalsRequest{
							{ID: "Ar5678", TeamID: "T5678", Status: api.AppsApprovalsRequestStatusApproved, DateCreated: mockRequestCreated},
							{ID: "Ar9012", TeamID: "T9012", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
						},
					}, nil)
			},
			Teardown: restoreRequest,
			ExpectedOutputs: []string{
				"teamtwo (T5678):",
				"Status:       approved",
				"T9012:",
				"Status:       pending",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListAppApprovalRequests", mock.Anything, "xoxp-selected", "A5678", []string(nil))
			},
		},
		"returns the error of a failed team selection": {
			CmdArgs: []string{"--app", "A5678"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequestWithoutProject(ctx, cm, cf)
				requestTeamSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, promptText string, promptConfig *prompts.PromptTeamSlackAuthConfig) (*types.SlackAuth, error) {
					return nil, slackerror.New(slackerror.ErrProcessInterrupted)
				}
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrProcessInterrupted),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"errors when the selected team has no token": {
			CmdArgs: []string{"--app", "A5678"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequestWithoutProject(ctx, cm, cf)
				requestTeamSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, promptText string, promptConfig *prompts.PromptTeamSlackAuthConfig) (*types.SlackAuth, error) {
					return &types.SlackAuth{TeamID: "T5678"}, nil
				}
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrCredentialsNotFound),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"errors without a project when an app environment is used": {
			CmdArgs: []string{"--app", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequestWithoutProject(ctx, cm, cf)
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrInvalidAppDirectory),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"returns the error of an interrupted app selection": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				requestAppSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, environment prompts.AppEnvironmentType, status prompts.AppInstallStatus, opts ...prompts.AppSelectOption) (prompts.SelectedApp, error) {
					return prompts.SelectedApp{}, slackerror.New(slackerror.ErrProcessInterrupted)
				}
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrProcessInterrupted),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"errors when the selected app is missing an ID": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				requestAppSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, environment prompts.AppEnvironmentType, status prompts.AppInstallStatus, opts ...prompts.AppSelectOption) (prompts.SelectedApp, error) {
					return prompts.SelectedApp{Auth: types.SlackAuth{Token: "xoxp-example"}}, nil
				}
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrAppNotFound),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListAppApprovalRequests")
			},
		},
		"returns the error of a failed lookup": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableRequest(ctx, cm, cf)
				cm.API.On("ListAppApprovalRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(api.AppsApprovalsRequestsListResult{}, slackerror.New(slackerror.ErrAPIFeatureNotEnabled))
			},
			Teardown:      restoreRequest,
			ExpectedError: slackerror.New(slackerror.ErrAPIFeatureNotEnabled),
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewRequestCommand(cf)
	})
}

func TestRequestFormat(t *testing.T) {
	tests := map[string]struct {
		TeamNames  map[string]string
		Requests   []api.AppsApprovalsRequest
		Expected   []string
		Unexpected []string
	}{
		"no request was made for the app": {
			Requests: []api.AppsApprovalsRequest{},
			Expected: []string{
				"App ID:       A1234",
				"You have not requested to install this app",
			},
		},
		"the app is named before the requests of each team": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
			},
			Expected: []string{
				"App ID:       A1234",
				"T1234",
				"Request ID:   Ar1234",
			},
		},
		"an open request omits the resolved timestamp": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
			},
			Expected: []string{
				"T1234",
				"Request ID:   Ar1234",
				"Status:       pending",
				"Requested:    2026-08-21",
			},
			Unexpected: []string{"Resolved:", "Cancelled by:", "without approval"},
		},
		"an approved request includes the resolved timestamp": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusApproved, DateCreated: mockRequestCreated, DateResolved: mockRequestResolved},
			},
			Expected: []string{
				"Status:       approved",
				"Requested:    2026-08-21",
				"Resolved:     2026-08-22",
			},
		},
		"a request cancelled by an admin of the team": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusCancelled, DateCreated: mockRequestCreated, DateResolved: mockRequestResolved, CancelledBy: api.AppsApprovalsRequestCancelledByAdmin},
			},
			Expected: []string{
				"Status:       cancelled",
				"Cancelled by: an admin",
			},
		},
		"a request withdrawn by the authenticated account": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusCancelled, DateCreated: mockRequestCreated, DateResolved: mockRequestResolved, CancelledBy: api.AppsApprovalsRequestCancelledByUser},
			},
			Expected: []string{"Cancelled by: you"},
		},
		"a request cancelled without an actor of its own": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusCancelled, DateCreated: mockRequestCreated, DateResolved: mockRequestResolved, CancelledBy: api.AppsApprovalsRequestCancelledBySystem},
			},
			Expected: []string{"Cancelled by: the system"},
		},
		"a denied request that the account can approve itself": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusDenied, CanSelfApprove: true, DateCreated: mockRequestCreated, DateResolved: mockRequestResolved},
			},
			Expected: []string{
				"Status:       denied",
				"You can install this app without approval. Please cancel the request.",
			},
		},
		"a request without a timestamp reports an unknown moment": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending},
			},
			Expected:   []string{"Requested:    unknown"},
			Unexpected: []string{"Resolved:"},
		},
		"an unrecognized status is reported without styles": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatus("escalated"), DateCreated: mockRequestCreated},
			},
			Expected: []string{"Status:       escalated"},
		},
		"an unrecognized cancellation actor is reported as named": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusCancelled, DateCreated: mockRequestCreated, CancelledBy: api.AppsApprovalsRequestCancelledBy("workflow")},
			},
			Expected: []string{"Cancelled by: workflow"},
		},
		"a searched team is titled by name when it is known": {
			TeamNames: map[string]string{"T1234": "teamone"},
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
				{ID: "Ar5678", TeamID: "T5678", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
			},
			Expected: []string{
				"teamone (T1234):",
				"T5678:",
			},
		},
		"requests are sorted by the team ID": {
			Requests: []api.AppsApprovalsRequest{
				{ID: "Ar5678", TeamID: "T5678", Status: api.AppsApprovalsRequestStatusCancelled, DateCreated: mockRequestCreated},
				{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusApproved, DateCreated: mockRequestCreated},
			},
			Expected: []string{
				"T1234",
				"Status:       approved",
				"T5678",
				"Status:       cancelled",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			formatted := strings.Join(FormatRequestSuccess("A1234", tc.TeamNames, tc.Requests), "\n")
			previous := -1
			for _, value := range tc.Expected {
				index := strings.Index(formatted, value)
				assert.Greater(t, index, previous, "expected %q to follow the preceding values", value)
				previous = index
			}
			for _, value := range tc.Unexpected {
				assert.NotContains(t, formatted, value)
			}
		})
	}
}

// TestRequestFormatOrdering checks that the requests of the caller are left in
// the order they were provided while output remains sorted by team
func TestRequestFormatOrdering(t *testing.T) {
	requests := []api.AppsApprovalsRequest{
		{ID: "Ar5678", TeamID: "T5678", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
		{ID: "Ar1234", TeamID: "T1234", Status: api.AppsApprovalsRequestStatusPending, DateCreated: mockRequestCreated},
	}

	formatted := strings.Join(FormatRequestSuccess("A1234", nil, requests), "\n")

	assert.Less(t, strings.Index(formatted, "Ar1234"), strings.Index(formatted, "Ar5678"))
	assert.Equal(t, "Ar5678", requests[0].ID, "expected the provided requests to remain unsorted")
	assert.Equal(t, "Ar1234", requests[1].ID, "expected the provided requests to remain unsorted")
}

// mockRequestWorkspaceIDs returns a count of unique workspace IDs
func mockRequestWorkspaceIDs(count int) []string {
	workspaceIDs := []string{}
	for i := range count {
		workspaceIDs = append(workspaceIDs, fmt.Sprintf("T%09d", i))
	}
	return workspaceIDs
}
