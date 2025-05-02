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

package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateApp_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  appManifestCreateMethod,
		ExpectedRequest: `{"manifest":{"display_information":{"name":"TestApp"}},"enable_distribution":true}`,
		Response:        `{"ok":true,"app_id":"A123"}`,
	})
	defer teardown()
	result, err := c.CreateApp(ctx, "token", types.AppManifest{
		DisplayInformation: types.DisplayInformation{
			Name: "TestApp",
		},
	}, /* enableDistribution= */ true)
	require.NoError(t, err)
	require.Equal(t, "A123", result.AppID)
}

func TestClient_CreateApp_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, appManifestCreateMethod, func(c *Client) error {
		_, err := c.CreateApp(ctx, "token", types.AppManifest{}, false)
		return err
	})
}

func TestClient_ExportAppManifest_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  appManifestExportMethod,
		ExpectedRequest: `{"app_id":"A0123456789"}`,
		Response:        `{"ok":true,"manifest":{"display_information":{"name":"example"}}}`,
	})
	defer teardown()
	result, err := c.ExportAppManifest(ctx, "token", "A0123456789")
	require.NoError(t, err)
	require.Equal(t, "example", result.Manifest.AppManifest.DisplayInformation.Name)
}

func TestClient_ExportAppManifest_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, appManifestExportMethod, func(c *Client) error {
		_, err := c.ExportAppManifest(ctx, "token", "A0000000001")
		return err
	})
}

func TestClient_UpdateApp_OK(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  appManifestUpdateMethod,
		ExpectedRequest: `{"manifest":{"display_information":{"name":"TestApp"},"outgoing_domains":[]},"app_id":"A123","force_update":true,"consent_breaking_changes":true}`,
		Response:        `{"ok":true,"app_id":"A123"}`,
	})
	defer teardown()
	result, err := c.UpdateApp(ctx, "token", "A123", types.AppManifest{
		DisplayInformation: types.DisplayInformation{
			Name: "TestApp",
		},
		OutgoingDomains: &[]string{},
	},
		/* forceUpdate= */ true,
		/* continueWithBreakingChanges= */ true)
	require.NoError(t, err)
	require.Equal(t, "A123", result.AppID)
}

func TestClient_UpdateApp_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, appManifestUpdateMethod, func(c *Client) error {
		_, err := c.UpdateApp(ctx, "token", "A123", types.AppManifest{},
			/* forceUpdate= */ true,
			/* continueWithBreakingChanges= */ true)
		return err
	})
}

func TestClient_UpdateApp_SchemaCompatibilityError(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appManifestUpdateMethod,
		Response:       `{"ok": false, "error": "err", "errors": [{"message": "schema_compatibility_error: Following datatstore(s) will be deleted after the deploy: tasks"}]}`,
	})
	defer teardown()
	_, err := c.UpdateApp(ctx, "token", "A123", types.AppManifest{},
		/* forceUpdate= */ true,
		/* continueWithBreakingChanges= */ true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "schema_compatibility_error")
}

func TestClient_ValidateAppManifest(t *testing.T) {
	// Error Fixtures
	var errInvalidManifestCode = "invalid_manifest"
	var errInvalidManifestDesc = "Illegal bot scopes found"

	// Error detail illegal_bot_scopes
	var errInvalidManifestDetailIllegalBotScopes = slackerror.ErrorDetail{
		Code:    "illegal_bot_scopes",
		Message: "Illegal bot scopes found 'thisisnotascope'",
		Pointer: "/oauth_config/scopes/bot",
	}
	var errorDetails slackerror.ErrorDetails
	errorDetails = append(errorDetails, errInvalidManifestDetailIllegalBotScopes)

	// Error detail connector_not_installed
	var errInvalidManifestDetailConnectorNotInstalled = slackerror.ErrorDetail{
		Code:    slackerror.ErrConnectorNotInstalled,
		Message: "Function `post_random_gif` from `Giphy` found in step `0` of the Workflow references Workflow Steps that are not installed in the Workspace.",
		Pointer: "/workflows/post_random_gif/steps/0",
	}

	var errorDetails2 slackerror.ErrorDetails
	errorDetails2 = append(errorDetails2, errInvalidManifestDetailConnectorNotInstalled)

	// Warning Fixtures
	var breakingChangeWarning = slackerror.Warning{
		Code:    "breaking_change",
		Message: "Required properties removed: ['id'].",
		Pointer: "/events/incident_created",
	}

	var warnings slackerror.Warnings
	warnings = append(warnings, breakingChangeWarning)

	// Setup tests
	type args struct {
		token    string
		appID    string
		manifest types.AppManifest
	}
	tests := []struct {
		name        string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		args        args
		want        ValidateAppManifestResult
		wantErr     bool
		wantErrVal  error
	}{
		{
			name: "handles an ok error ",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				result := `{"ok": true, "errors":[]}`
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			},
			want: ValidateAppManifestResult{nil},
		},
		{
			name: "returns an error and error details on invalid_manifest",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				result := fmt.Sprintf(
					`{"ok":false,"error":"%s","slack_cli_error_description":"%s","errors":[{"code":"%s","message":"%s","pointer":"%s"}]}`,
					errInvalidManifestCode,
					errInvalidManifestDesc,
					errInvalidManifestDetailIllegalBotScopes.Code,
					errInvalidManifestDetailIllegalBotScopes.Message,
					errInvalidManifestDetailIllegalBotScopes.Pointer,
				)
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			},
			want:       ValidateAppManifestResult{nil},
			wantErr:    true,
			wantErrVal: slackerror.NewAPIError(errInvalidManifestCode, errInvalidManifestDesc, errorDetails, appManifestValidateMethod),
		},
		{
			name: "returns warning with breaking_change",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				result := fmt.Sprintf(
					`{"ok":true,"errors":[],"warnings":[{"code":"%s","message":"%s","pointer":"%s"}]}`,
					breakingChangeWarning.Code,
					breakingChangeWarning.Message,
					breakingChangeWarning.Pointer,
				)
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			},
			want: ValidateAppManifestResult{warnings},
		},
		// This shouldn't happen right now, but adding a test to make sure that if we start sending back warnings AND
		// errors nothing breaks and nothing unexpected happens - we should still return JUST the error and no warning
		{
			name: "returns error on invalid_manifest and warnings with breaking_change",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				result := fmt.Sprintf(
					`{"ok":true,"error":"%s","slack_cli_error_description":"%s","errors":[{"code":"%s","message":"%s","pointer":"%s"}],"warnings":[{"code":"%s","message":"%s","pointer":"%s"}]}`,
					errInvalidManifestCode,
					errInvalidManifestDesc,
					errInvalidManifestDetailIllegalBotScopes.Code,
					errInvalidManifestDetailIllegalBotScopes.Message,
					errInvalidManifestDetailIllegalBotScopes.Pointer,
					breakingChangeWarning.Code,
					breakingChangeWarning.Message,
					breakingChangeWarning.Pointer,
				)
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			},
			want:       ValidateAppManifestResult{nil},
			wantErr:    true,
			wantErrVal: slackerror.NewAPIError(errInvalidManifestCode, errInvalidManifestDesc, errorDetails, appManifestValidateMethod),
		},
		{
			name: "returns an error invalid_manifest when error detail is due to connector not being installed",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				result := fmt.Sprintf(
					`{"ok":false,"error":"%s","errors":[{"code":"%s","message":"%s","pointer":"%s"}]}`,
					errInvalidManifestCode,
					errInvalidManifestDetailConnectorNotInstalled.Code,
					errInvalidManifestDetailConnectorNotInstalled.Message,
					errInvalidManifestDetailConnectorNotInstalled.Pointer,
				)
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			},
			want:       ValidateAppManifestResult{nil},
			wantErr:    true,
			wantErrVal: slackerror.NewAPIError(errInvalidManifestCode, "", errorDetails2, appManifestValidateMethod),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			var ts = httptest.NewServer(http.HandlerFunc(tt.handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			got, err := c.ValidateAppManifest(ctx, tt.args.token, tt.args.manifest, tt.args.appID)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrVal.Error())
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_GetPresignedS3PostParams_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  appGeneratePresignedPostMethod,
		ExpectedRequest: `{"app_id":"A123"}`,
		Response:        `{"ok":true,"url":"example.com/upload","file_name":"foo.tar.gz","fields":{"X-Amz-Credential":"cred"}}`,
	})
	defer teardown()
	result, err := c.GetPresignedS3PostParams(ctx, "token", "A123")
	require.NoError(t, err)
	require.Equal(t, "foo.tar.gz", result.FileName)
	require.Equal(t, "example.com/upload", result.URL)
	require.Equal(t, "cred", result.Fields.AmzCredentials)
}

func TestClient_GetPresignedS3PostParams_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, appGeneratePresignedPostMethod, func(c *Client) error {
		_, err := c.GetPresignedS3PostParams(ctx, "token", "A123")
		return err
	})
}

func TestClient_CertifiedAppInstall(t *testing.T) {
	tests := []struct {
		name       string
		resultJSON string
		wantErr    bool
		err        string
	}{
		{
			name:       "OK result",
			resultJSON: `{"ok":true}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, appCertifiedInstallMethod)
				expectedJSON := `{"app_id":"A123"}`
				payload, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, expectedJSON, string(payload))
				result := tt.resultJSON
				_, err = fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			mockAppID := "A123"
			// execute
			_, err := c.CertifiedAppInstall(ctx, "token", mockAppID)

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.err,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_InstallApp(t *testing.T) {

	var setup = func(t *testing.T) (context.Context, *iostreams.IOStreamsMock) {
		// Mocks
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		return ctx, ioMock
	}

	tests := []struct {
		name       string
		resultJSON string
		wantErr    bool
		err        string
	}{
		{
			name:       "OK result",
			resultJSON: `{"ok":true}`,
		}, {
			name:       "Error result",
			resultJSON: `{"ok":false,"error":"invalid_app_id"}`,
			wantErr:    true,
			err:        "invalid_app_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, ioMock := setup(t)

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, appDeveloperInstallMethod)
				expectedJSON := `{"app_id":"A123"}`
				payload, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, expectedJSON, string(payload))
				result := tt.resultJSON
				_, err = fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			mockApp := types.App{
				AppID:      "A123",
				TeamID:     "T123",
				TeamDomain: "mock",
			}
			// execute
			_, _, err := c.DeveloperAppInstall(ctx, ioMock, "token", mockApp, []string{}, []string{}, "", false)

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.err,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_UninstallApp(t *testing.T) {
	tests := []struct {
		name       string
		resultJSON string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "OK result",
			resultJSON: `{"ok":true}`,
		},
		{
			name:       "Error result",
			resultJSON: `{"ok":false,"error":"invalid_app_id"}`,
			wantErr:    true,
			errMessage: "invalid_app_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, appDeveloperUninstallMethod)
				expectedJSON := `{"app_id":"A123","team_id":"T123"}`
				payload, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, expectedJSON, string(payload))
				result := tt.resultJSON
				_, err = fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			err := c.UninstallApp(ctx, "token", "A123", "T123")

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_DeleteApp(t *testing.T) {
	tests := []struct {
		name       string
		resultJSON string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "OK result",
			resultJSON: `{"ok":true}`,
		},
		{
			name:       "Error result",
			resultJSON: `{"ok":false,"error":"invalid_app_id"}`,
			wantErr:    true,
			errMessage: "invalid_app_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, appDeleteMethod)
				expectedJSON := `{"app_id":"A123"}`
				payload, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, expectedJSON, string(payload))
				result := tt.resultJSON
				_, err = fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			err := c.DeleteApp(ctx, "token", "A123")

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_DeveloperAppInstall_RequestAppApproval(t *testing.T) {
	tests := []struct {
		name                string
		app                 types.App
		orgGrantWorkspaceID string
		teamID              string
		requestJSON         string
		wantErr             bool
		errMessage          string
	}{
		{
			name: `Standalone workspace, AAA is requested \
			(workspace ID passed into apps.approvals.requests.create)`,
			app:                 types.App{AppID: "A1234", TeamID: "T1234"},
			orgGrantWorkspaceID: "",
			teamID:              "T1234",
			requestJSON:         `{"app":"A1234","reason":"This request has been automatically generated according to project environment settings.","team_id":"T1234"}`,
		},
		{
			name: `User tried to install to a single workspace in an org, AAA is requested \
			(workspace ID passed into apps.approvals.requests.create)`,
			app:                 types.App{AppID: "A1234", EnterpriseID: "E1234", TeamID: "E1234"},
			orgGrantWorkspaceID: "T1234",
			teamID:              "T1234",
			requestJSON:         `{"app":"A1234","reason":"This request has been automatically generated according to project environment settings.","team_id":"T1234"}`,
		},
		{
			name: `User tried to install to all workspaces in an org, AAA is requested \
				(no team_id passed into apps.approvals.requests.create so it will default to creating a request for the auth team ie. the org)`,
			app:                 types.App{AppID: "A1234", EnterpriseID: "E1234", TeamID: "E1234"},
			orgGrantWorkspaceID: "all",
			teamID:              "E1234",
			requestJSON:         `{"app":"A1234","reason":"This request has been automatically generated according to project environment settings."}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {

				if strings.Contains(r.URL.Path, appDeveloperInstallMethod) {
					result := fmt.Sprintf(
						`{"ok":false,"error":"%s","team_id":"%s"}`,
						slackerror.ErrAppApprovalRequestEligible,
						tt.teamID,
					)
					_, err := fmt.Fprintln(w, result)
					require.NoError(t, err)
				}

				if strings.Contains(r.URL.Path, appApprovalRequestCreateMethod) {
					expectedJSON := tt.requestJSON
					payload, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, expectedJSON, string(payload))
					_, err = fmt.Fprintln(w, `{"ok":true}`)
					require.NoError(t, err)
				}
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			iostreamMock := iostreams.NewIOStreamsMock(&config.Config{}, &slackdeps.FsMock{}, &slackdeps.OsMock{})
			iostreamMock.On("PrintTrace", mock.Anything, mock.Anything, mock.Anything).Return()

			// execute
			_, _, err := c.DeveloperAppInstall(ctx, iostreamMock, "token", tt.app, []string{}, []string{}, tt.orgGrantWorkspaceID, true)
			require.NoError(t, err)
		})
	}
}
