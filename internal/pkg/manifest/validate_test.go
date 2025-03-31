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

package manifest

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ManifestValidate_GetManifestLocal_Error(t *testing.T) {
	ctx, clients, _, log, appMock, authMock := setupCommonMocks()

	// Mock the manifest to return error on get
	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything).Return(types.SlackYaml{}, slackerror.New("An error"))
	clients.AppClient().Manifest = manifestMock

	// Test
	logEvent, _, err := ManifestValidate(ctx, clients, log, appMock, authMock.Token)

	assert.Error(t, err)
	assert.Nil(t, logEvent)
}

func Test_ManifestValidate_Success(t *testing.T) {
	t.Run("should return success when no errors or warnings", func(t *testing.T) {

		ctx, clients, clientsMock, log, appMock, authMock := setupCommonMocks()

		// Mock manifest validation api result with no error
		clientsMock.ApiInterface.On("ValidateAppManifest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ValidateAppManifestResult{}, nil)

		// Test
		logEvent, _, err := ManifestValidate(ctx, clients, log, appMock, authMock.Token)

		assert.NoError(t, err)
		assert.Equal(t, "success", logEvent.Name)
	})
}

func Test_ManifestValidate_Warnings(t *testing.T) {
	t.Run("should return warnings", func(t *testing.T) {

		ctx, clients, clientsMock, log, appMock, authMock := setupCommonMocks()

		// Mock manifest validation api result with no error
		clientsMock.ApiInterface.On("ValidateAppManifest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ValidateAppManifestResult{
			Warnings: slackerror.Warnings{
				slackerror.Warning{
					Code: "dummy warning",
				},
			},
		}, nil)

		// Test
		_, warnings, err := ManifestValidate(ctx, clients, log, appMock, authMock.Token)

		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, warnings[0].Code, "dummy warning")
		assert.Equal(t, 1, len(warnings))
	})
}

func Test_ManifestValidate_Error(t *testing.T) {
	t.Run("should return error when there are errors", func(t *testing.T) {

		ctx, clients, clientsMock, log, appMock, authMock := setupCommonMocks()

		// Mock manifest validation api result with an error
		clientsMock.ApiInterface.On("ValidateAppManifest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			api.ValidateAppManifestResult{},
			slackerror.New("a dummy error").WithDetails(slackerror.ErrorDetails{
				slackerror.ErrorDetail{
					Code: "dummy_error_detail_code",
				},
			}))

		// Test
		logEvent, _, err := ManifestValidate(ctx, clients, log, appMock, authMock.Token)

		assert.Nil(t, logEvent)
		assert.Error(t, err)
	})
}

func Test_ManifestValidate_Error_ErrConnectorNotInstalled(t *testing.T) {
	t.Run("should try to install connector apps when there are related errors", func(t *testing.T) {
		ctx, clients, clientsMock, log, appMock, authMock := setupCommonMocks()

		// Mock manifest validation api result with an error and error details
		clientsMock.ApiInterface.On("ValidateAppManifest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ValidateAppManifestResult{
			Warnings: nil,
		}, slackerror.New("a dummy error").WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorNotInstalled,
				RelatedComponent: "A12345",
				Message:          "",
				Pointer:          "",
			},
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorNotInstalled,
				RelatedComponent: "A56789",
				Message:          "",
				Pointer:          "",
			},
		}))

		// Mock CertifiedAppInstall method success
		clientsMock.ApiInterface.On("CertifiedAppInstall", mock.Anything, authMock.Token, mock.Anything).Return(api.CertifiedInstallResult{}, nil)

		// Test
		_, _, err := ManifestValidate(ctx, clients, log, appMock, authMock.Token)

		// Since we've mocked the ValidateAppManifest call to return an error, we still expect this method to return an error
		// despite a successful CertifiedAppInstall call. That is realistic given that a manifest can error for other reasons
		// even after successful CertifiedAppInstall call.
		assert.Error(t, err)

		clientsMock.ApiInterface.AssertCalled(t, "CertifiedAppInstall", mock.Anything, authMock.Token, "A12345")
		clientsMock.ApiInterface.AssertCalled(t, "CertifiedAppInstall", mock.Anything, authMock.Token, "A56789")
		clientsMock.ApiInterface.AssertNumberOfCalls(t, "CertifiedAppInstall", 2)
		clientsMock.ApiInterface.AssertNumberOfCalls(t, "ValidateAppManifest", 2)
	})
}

func Test_HandleConnectorApprovalRequired(t *testing.T) {

	test_reason := "GIVE IT TO ME!"
	t.Run("should send request to approve connector", func(t *testing.T) {
		ctx, clients, clientsMock, _, _, authMock := setupCommonMocks()
		testErr := slackerror.New("a dummy error").WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorApprovalRequired,
				RelatedComponent: "A12345",
				Message:          "Step `connector_step` from the `connector` Connector found in step `0` of the Workflow",
				Pointer:          "",
			},
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorApprovalRequired,
				RelatedComponent: "A56789",
				Message:          "Step `connector_step` from the `connector` Connector found in step `0` of the Workflow",
				Pointer:          "",
			},
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorNotInstalled,
				RelatedComponent: "A1234",
				Message:          "",
				Pointer:          "",
			},
		})

		clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		clientsMock.IO.On("InputPrompt", mock.Anything, mock.Anything, mock.Anything).Return(test_reason, nil)
		clientsMock.ApiInterface.On("RequestAppApproval", mock.Anything, authMock.Token, mock.Anything, mock.Anything, test_reason, mock.Anything, mock.Anything).Return(api.AppsApprovalsRequestsCreateResult{}, nil)

		// Test
		err := HandleConnectorApprovalRequired(ctx, clients, authMock.Token, testErr)

		assert.NoError(t, err)
		clientsMock.ApiInterface.AssertCalled(t, "RequestAppApproval", mock.Anything, authMock.Token, "A12345", mock.Anything, test_reason, mock.Anything, mock.Anything)
		clientsMock.ApiInterface.AssertCalled(t, "RequestAppApproval", mock.Anything, authMock.Token, "A56789", mock.Anything, test_reason, mock.Anything, mock.Anything)
		clientsMock.ApiInterface.AssertNumberOfCalls(t, "RequestAppApproval", 2)
		assert.Equal(t, len(testErr.Details), 1)
	})

	t.Run("should not send request to approve connector if user refuses", func(t *testing.T) {
		ctx, clients, clientsMock, _, _, authMock := setupCommonMocks()
		testErr := slackerror.New("a dummy error").WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorApprovalRequired,
				RelatedComponent: "A12345",
				Message:          "Step `connector_step` from the `connector` Connector found in step `0` of the Workflow",
				Pointer:          "",
			},
		})

		clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		// Test
		err := HandleConnectorApprovalRequired(ctx, clients, authMock.Token, testErr)

		assert.NoError(t, err)
		clientsMock.ApiInterface.AssertNumberOfCalls(t, "RequestAppApproval", 0)
		assert.Equal(t, len(testErr.Details), 0)
	})

	t.Run("should return error if request RequestAppApproval fails", func(t *testing.T) {
		ctx, clients, clientsMock, _, _, authMock := setupCommonMocks()
		testErr := slackerror.New("a dummy error").WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{
				Code:             slackerror.ErrConnectorApprovalRequired,
				RelatedComponent: "A12345",
				Message:          "Step `connector_step` from the `connector` Connector found in step `0` of the Workflow",
				Pointer:          "",
			},
		})

		clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		clientsMock.IO.On("InputPrompt", mock.Anything, mock.Anything, mock.Anything).Return(test_reason, nil)
		clientsMock.ApiInterface.On("RequestAppApproval", mock.Anything, authMock.Token, mock.Anything, mock.Anything, test_reason, mock.Anything, mock.Anything).Return(api.AppsApprovalsRequestsCreateResult{}, slackerror.New("dummy error"))

		// Test
		err := HandleConnectorApprovalRequired(ctx, clients, authMock.Token, testErr)

		assert.Error(t, err)
		clientsMock.ApiInterface.AssertNumberOfCalls(t, "RequestAppApproval", 1)
	})
}

// Setup
func setupCommonMocks() (ctx context.Context, clients *shared.ClientFactory, clientsMock *shared.ClientsMock, log *logger.Logger, mockApp types.App, mockAuth types.SlackAuth) {
	// Create mocks
	clientsMock = shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	// Create clients that is mocked for testing
	clients = shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	ctx = context.Background()

	// Mock valid auth session
	clientsMock.ApiInterface.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)

	// Mock the manifest
	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything).Return(types.SlackYaml{}, nil)
	clients.AppClient().Manifest = manifestMock

	// Setup logger
	log = &logger.Logger{}
	log.Data = logger.LogData{}

	// Create a dummy app
	mockApp = types.App{
		AppID:        "A123",
		EnterpriseID: "E123",
		TeamID:       "E123",
		UserID:       "U123",
		IsDev:        false,
		TeamDomain:   "workspace",
	}

	mockAuth = types.SlackAuth{
		Token:      "xoxp.xoxe-1234",
		TeamDomain: "workspace",
	}

	return ctx, clients, clientsMock, log, mockApp, mockAuth
}
