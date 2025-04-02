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

package auth

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthWithToken(t *testing.T) {
	os := slackdeps.NewOsMock()
	os.AddDefaultMocks()
	fs := slackdeps.NewFsMock()
	config := config.NewConfig(fs, os)

	tests := map[string]struct {
		token    string
		response string
		err      slackerror.Error
		expected types.SlackAuth
	}{
		"attempt to resolve the empty token": {
			token:    "",
			response: `{"ok":false,"error":"not_authed"}`,
			err:      slackerror.Error{Code: "not_authed"},
			expected: types.SlackAuth{},
		},
		"resolve the example authentication": {
			token:    "xoxp-example-010101",
			response: `{"ok":true,"url":"https://slundosip.slack.com","team":"slundos","user":"jartab","team_id":"T2244224488","user_id":"U6543210000","is_enterprise_install":false}`,
			err:      slackerror.Error{},
			expected: types.SlackAuth{
				Token:      "xoxp-example-010101",
				TeamDomain: "slundosip",
				TeamID:     "T2244224488",
				UserID:     "U6543210000",
			},
		},
		"resolve an enterprise authentication": {
			token:    "xoxp-enterpriser-44",
			response: `{"ok":true,"url":"https://starship.slack.com","team":"starship","user":"treker","team_id":"T23","user_id":"U9000","enterprise_id":"E99","is_enterprise_install":true}`,
			err:      slackerror.Error{},
			expected: types.SlackAuth{
				Token:               "xoxp-enterpriser-44",
				TeamDomain:          "starship",
				TeamID:              "T23",
				UserID:              "U9000",
				EnterpriseID:        "E99",
				IsEnterpriseInstall: true,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			apic, teardown := api.NewFakeClient(t, api.FakeClientParams{
				ExpectedMethod:  "auth.test",
				ExpectedRequest: fmt.Sprintf("token=%s", tt.token),
				Response:        tt.response,
			})
			defer teardown()
			appc := app.NewClient(apic, config, fs, os)
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
			ioMock.AddDefaultMocks()
			c := NewClient(apic, appc, config, ioMock, fs)

			auth, err := c.AuthWithToken(ctx, tt.token)
			auth.LastUpdated = time.Time{} // ignore time for this test

			if tt.err.Code != "" {
				assert.Error(t, err, slackerror.New(slackerror.ErrNotAuthed))
				assert.Equal(t, tt.expected, auth)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, auth)
			}
		})
	}
}

func Test_AuthGettersAndSetters(t *testing.T) {
	// Mock auth
	authMockA := types.SlackAuth{
		Token:      "xoxe-xoxp-",
		TeamDomain: "aaa",
		TeamID:     "T123456789A",
		UserID:     "Uxxxx",
	}
	authMockB := types.SlackAuth{
		Token:      "xoxe-xoxp-",
		TeamDomain: "bbb",
		TeamID:     "T123456789B",
		UserID:     "Uxxxx",
	}

	fakeAuths := []types.SlackAuth{
		authMockA,
		authMockB,
	}

	var setup = func(t *testing.T) (context.Context, *Client) {
		// Mocks
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)

		// Set up fake auths
		for _, auth := range fakeAuths {
			_, _, err := authClient.SetAuth(ctx, auth)
			require.NoError(t, err)
		}

		return ctx, authClient
	}
	t.Run("SetAuth properly sets auths and getters return expected values", func(t *testing.T) {
		ctx, authClient := setup(t)

		// Should be able to retrieve auths
		allAuthsMap, err := authClient.auths(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(allAuthsMap))

		// Should be able to retrieve auths as a slice
		allAuths, err := authClient.auths(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(allAuths))

		// Should be able to find each auth from the auths map
		_, authAExists := allAuthsMap[authMockA.TeamID]
		require.True(t, authAExists)

		// Should be able to find each auth from the auths map
		_, authBExists := allAuthsMap[authMockB.TeamID]
		require.True(t, authBExists)

		// Should be able to find each auth via teamDomain
		authMockARes, err := authClient.AuthWithTeamID(ctx, authMockA.TeamID)
		require.NoError(t, err)
		require.Equal(t, authMockA.TeamID, authMockARes.TeamID)

		// Should be able to find each auth via teamDomain
		authMockBRes, err := authClient.AuthWithTeamID(ctx, authMockB.TeamID)
		require.NoError(t, err)
		require.Equal(t, authMockB.TeamID, authMockBRes.TeamID)

		// Should be able to find each  via team ID
		authMockARes, err = authClient.AuthWithTeamID(ctx, authMockA.TeamID)
		require.NoError(t, err)
		require.Equal(t, authMockA.TeamDomain, authMockARes.TeamDomain)

		// Should be able to find each auth via team ID
		authMockBRes, err = authClient.AuthWithTeamID(ctx, authMockB.TeamID)
		require.NoError(t, err)
		require.Equal(t, authMockB.TeamDomain, authMockBRes.TeamDomain)
	})

	t.Run("DeleteAuth properly deletes an auth", func(t *testing.T) {
		ctx, authClient := setup(t)

		// Should be able to retrieve auths
		allAuths, err := authClient.Auths(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(allAuths))

		// Should be able to delete a single auth
		authToDelete := allAuths[0]
		deletedAuth, err := authClient.DeleteAuth(ctx, authToDelete)
		require.NoError(t, err)
		updatedAuths, _ := authClient.auths(ctx)
		require.Equal(t, 1, len(updatedAuths))

		// deletedAuth should not be findable anymore
		_, err = authClient.AuthWithTeamID(ctx, deletedAuth.TeamID)
		require.Error(t, err)
	})

	t.Run("invalid JSON is caught with a meaningful error", func(t *testing.T) {
		ctx, authClient := setup(t)
		dir, err := authClient.config.SystemConfig.SlackConfigDir(ctx)
		require.NoError(t, err)
		path := filepath.Join(dir, credentialsFileName)
		err = afero.WriteFile(authClient.fs, path, []byte("{"), 0o600)
		require.NoError(t, err)
		_, err = authClient.auths(ctx)
		require.Error(t, err)
		assert.Equal(t, err.(*slackerror.Error).Code, slackerror.ErrUnableToParseJson)
	})
}

func Test_AuthsRotation(t *testing.T) {
	// Setup tests
	var setup = func(t *testing.T) (context.Context, *Client) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)
		return ctx, authClient
	}

	t.Run("no rotation needed for any of the auths", func(t *testing.T) {
		// Setup
		ctx, authClient := setup(t)
		timeNow := int(time.Now().Unix())
		oneHourLater := timeNow + 60*60

		// Fixtures
		_defaultDevApiClientHost := defaultDevApiClientHost
		_defaultProdApiClientHost := defaultProdApiClientHost
		authATeamDomain := "workspace-a"
		authATeamId := "T123456789A"
		authBTeamDomain := "workspace-b"
		authBTeamId := "T123456789B"
		authA := types.SlackAuth{
			ApiHost:    &_defaultProdApiClientHost,
			ExpiresAt:  oneHourLater,
			TeamDomain: authATeamDomain,
			TeamID:     authATeamId,
		}
		authB := types.SlackAuth{
			ApiHost:    &_defaultDevApiClientHost,
			ExpiresAt:  oneHourLater,
			TeamDomain: authBTeamDomain,
			TeamID:     authBTeamId,
		}
		auths := types.AuthByTeamDomain{
			authATeamDomain: authA,
			authBTeamDomain: authB,
		}
		_, err := authClient.setAuths(ctx, auths)
		if err != nil {
			require.Fail(t, "should not return error when saving auths")
		}

		// setup api client
		authClient.api = api.NewClient(&http.Client{}, defaultProdApiClientHost, authClient.io)

		// call the function
		updatedAuths, err := authClient.auths(ctx)

		// Assertions
		require.Equal(t, authA, updatedAuths[authATeamId], "should return the same auth")
		require.Equal(t, authB, updatedAuths[authBTeamId], "should return the same auth")
		require.NoError(t, err, "Should not return an error when the contents of credentials are valid")
	})

	t.Run("token rotation needed for at least one of the auths", func(t *testing.T) {
		// Setup
		ctx, authClient := setup(t)
		timeNow := int(time.Now().Unix())
		oneHourLater := timeNow + 60*60
		fiveMinutesAgo := timeNow - 60*5

		// setup a fake api client since we don't want to make actual api calls
		mockApiClient, teardown := api.NewFakeClient(t, api.FakeClientParams{
			ExpectedMethod:  "tooling.tokens.rotate",
			ExpectedRequest: `refresh_token=valid-refresh-token`,
			Response:        fmt.Sprintf(`{"ok":true,"token": "new-token", "exp": %d, "refresh_token": "new-valid-refresh-token"}`, oneHourLater),
		})
		defer teardown()

		authClient.api = mockApiClient
		fakeApiHost := mockApiClient.Host()

		// Fixtures
		_defaultProdApiClientHost := defaultProdApiClientHost
		authATeamDomain := "workspace-a"
		authATeamId := "T123456789A"
		authBTeamDomain := "workspace-b"
		authBTeamId := "T123456789B"

		workspaceAuthA := types.SlackAuth{
			ApiHost:    &_defaultProdApiClientHost,
			Token:      "goodToken",
			ExpiresAt:  oneHourLater,
			TeamDomain: authATeamDomain,
			TeamID:     authATeamId,
		}

		// use the fake api host as the Api host so that we don't end up making actual api calls
		workspaceAuthB := types.SlackAuth{
			ApiHost:      &fakeApiHost,
			Token:        "expiredToken",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    fiveMinutesAgo,
			TeamDomain:   authBTeamDomain,
			TeamID:       authBTeamId,
		}

		auths := types.AuthByTeamDomain{
			authATeamDomain: workspaceAuthA,
			authBTeamDomain: workspaceAuthB,
		}
		_, err := authClient.setAuths(ctx, auths)
		if err != nil {
			require.Fail(t, "should not return error when saving auths")
		}
		apiHostBefore := authClient.api.Host()     // track the api host before we call User Auths
		updatedAuths, err := authClient.auths(ctx) // call the function
		apiHostAfter := authClient.api.Host()      // track the api host after the function runs.

		expectedWorkspaceAuthB := types.SlackAuth{ApiHost: &fakeApiHost, Token: "new-token", RefreshToken: "new-valid-refresh-token", ExpiresAt: oneHourLater}

		// Assertions
		require.Equal(t, workspaceAuthA, updatedAuths[authATeamId], "should return the same auth")
		require.Equal(t, expectedWorkspaceAuthB.Token, updatedAuths[authBTeamId].Token, "should return the renewed auth")
		require.Equal(t, expectedWorkspaceAuthB.RefreshToken, updatedAuths[authBTeamId].RefreshToken, "should return the renewed auth")
		require.Equal(t, expectedWorkspaceAuthB.ExpiresAt, updatedAuths[authBTeamId].ExpiresAt, "should return the renewed auth")
		require.Equal(t, apiHostBefore, apiHostAfter, "api host before and after should be the same")
		require.NoError(t, err, "Should not return an error when the contents of credentials are valid")
	})

	t.Run("token rotation returns an error", func(t *testing.T) {
		// Setup
		ctx, authClient := setup(t)
		timeNow := int(time.Now().Unix())
		oneHourLater := timeNow + 60*60
		fiveMinutesAgo := timeNow - 60*5

		// setup a fake api client since we don't want to make actual api calls
		mockApiClient, teardown := api.NewFakeClient(t, api.FakeClientParams{
			ExpectedMethod:  "tooling.tokens.rotate",
			ExpectedRequest: `refresh_token=valid-refresh-token`,
			Response:        `{"ok":false,"error": "invalid_auth"}`,
		})
		defer teardown()

		authClient.api = mockApiClient
		fakeApiHost := mockApiClient.Host()

		// Fixtures
		_defaultProdApiClientHost := defaultProdApiClientHost
		authATeamDomain := "workspace-a"
		authATeamId := "T123456789A"
		authBTeamDomain := "workspace-b"
		authBTeamId := "T123456789B"

		workspaceAuthA := types.SlackAuth{
			ApiHost:    &_defaultProdApiClientHost,
			Token:      "goodToken",
			ExpiresAt:  oneHourLater,
			TeamDomain: authATeamDomain,
			TeamID:     authATeamId,
		}

		// use the fake api host as the Api host so that we don't end up making actual api calls
		workspaceAuthB := types.SlackAuth{
			ApiHost:      &fakeApiHost,
			Token:        "expiredToken",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    fiveMinutesAgo,
			TeamDomain:   authBTeamDomain,
			TeamID:       authBTeamId,
		}

		auths := types.AuthByTeamDomain{
			authATeamId: workspaceAuthA,
			authBTeamId: workspaceAuthB,
		}
		_, err := authClient.setAuths(ctx, auths)
		if err != nil {
			require.Fail(t, "should not return error when saving auths")
		}

		updatedAuths, err := authClient.auths(ctx) // call the function

		// Assertions
		require.Equal(t, workspaceAuthA, updatedAuths[authATeamId], "should return the same auth")

		// because the token rotation results in an error we expect the old stuff back
		require.Equal(t, workspaceAuthB, updatedAuths[authBTeamId], "should return the same auth")

		require.Equal(t, len(auths), len(updatedAuths), "we expect the same number of auths even if token rotation failed for one")
		require.NoError(t, err, "Should not return an error when the contents of credentials are valid")
	})
}

func Test_Auths(t *testing.T) {
	var setup = func(t *testing.T) (context.Context, *Client) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)
		return ctx, authClient
	}

	t.Run("Should return Auths as a slice", func(t *testing.T) {
		// Setup
		ctx, authClient := setup(t)
		timeNow := int(time.Now().Unix())
		oneHourLater := timeNow + 60*60

		// Fixtures
		_defaultDevApiClientHost := defaultDevApiClientHost
		_defaultProdApiClientHost := defaultProdApiClientHost
		authATeamDomain := "workspace-a"
		authATeamId := "T123456789A"
		authBTeamDomain := "workspace-b"
		authBTeamId := "T123456789B"
		authA := types.SlackAuth{
			ApiHost:    &_defaultProdApiClientHost,
			ExpiresAt:  oneHourLater,
			TeamDomain: authATeamDomain,
			TeamID:     authATeamId,
		}
		authB := types.SlackAuth{
			ApiHost:    &_defaultDevApiClientHost,
			ExpiresAt:  oneHourLater,
			TeamDomain: authBTeamDomain,
			TeamID:     authBTeamId,
		}
		auths := types.AuthByTeamDomain{
			authATeamDomain: authA,
			authBTeamDomain: authB,
		}
		_, err := authClient.setAuths(ctx, auths)
		if err != nil {
			require.Fail(t, "should not return error when saving auths")
		}

		// setup api client
		authClient.api = api.NewClient(&http.Client{}, defaultProdApiClientHost, authClient.io)

		// call the function
		updatedAuths, err := authClient.Auths(ctx)

		require.NoError(t, err)
		require.Equal(t, 2, len(updatedAuths))
	})
}

func Test_migrateToAuthByTeamID(t *testing.T) {
	var setup = func(t *testing.T) (context.Context, *Client) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)
		return ctx, authClient
	}
	ctx, authClient := setup(t)

	t.Run("Always returns auths keyed by team_id", func(t *testing.T) {
		// Setup mock auths
		mockAuths := types.AuthByTeamDomain{
			"domain": {
				TeamDomain: "domain",
				TeamID:     "T123456789A",
			},
			"domain-1": {
				TeamDomain: "domain-1",
				TeamID:     "T123456789B",
			},
			"T123456789C": {
				TeamDomain: "domain-2",
				TeamID:     "T123456789C",
			},
			"E123456789D": {
				TeamDomain: "domain-3",
				TeamID:     "E123456789D",
			},
		}
		updatedAuths, hasUpdate := authClient.migrateToAuthByTeamID(ctx, mockAuths)
		require.True(t, hasUpdate)

		for k := range updatedAuths {
			if !authClient.isTeamID(k) {
				require.FailNow(t, "all user auths must be keyed by team_id")
			}
		}

		// all auths must be present
		require.Equal(t, 4, len(mockAuths))
		for _, auth := range mockAuths {
			teamId := auth.TeamID
			if _, exists := updatedAuths[teamId]; !exists {
				require.FailNow(t, "missing auth after update")
			}
		}
	})
}

func Test_SetSelectedAuth(t *testing.T) {
	var setup = func(t *testing.T) (context.Context, *Client, *slackdeps.OsMock) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)
		return ctx, authClient, osMock
	}
	mockAPIHost := "dev.slack.com"
	tests := map[string]struct {
		auth            types.SlackAuth
		expectedAPIHost string
		expectedAPIURL  string
	}{
		"associated authentication configurations are made": {
			auth: types.SlackAuth{
				TeamID:  "T001",
				ApiHost: &mockAPIHost,
			},
			expectedAPIHost: fmt.Sprintf("https://%s", mockAPIHost),
			expectedAPIURL:  fmt.Sprintf("https://%s/api/", mockAPIHost),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, authClient, osMock := setup(t)
			authClient.SetSelectedAuth(ctx, tt.auth, authClient.config, osMock)
			assert.Equal(t, authClient.config.TeamFlag, tt.auth.TeamID)
			assert.Equal(t, authClient.config.ApiHostResolved, tt.expectedAPIHost)
			osMock.AssertCalled(t, "Setenv", "SLACK_API_URL", tt.expectedAPIURL)
		})
	}
}

func Test_IsApiHostSlackDev(t *testing.T) {

	var setup = func(t *testing.T) (context.Context, *Client) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		config := config.NewConfig(fsMock, osMock)
		ioMock := iostreams.NewIOStreamsMock(config, fsMock, osMock)
		ioMock.AddDefaultMocks()
		authClient := NewClient(nil, nil, config, ioMock, fsMock)
		return ctx, authClient
	}
	_, authClient := setup(t)

	mockHostName := ""
	mockHostName1 := "notAUrl"
	mockHostName2 := "https://doesnothaveprefix.com"
	mockHostName3 := "https://dev1234.slack.com"
	mockHostName4 := "https://dev.slack.com"

	t.Run("Should validate api host slack dev", func(t *testing.T) {
		res := authClient.IsApiHostSlackDev(mockHostName)
		require.False(t, res)

		res = authClient.IsApiHostSlackDev(mockHostName1)
		require.False(t, res)

		res = authClient.IsApiHostSlackDev(mockHostName2)
		require.False(t, res)

		res = authClient.IsApiHostSlackDev(mockHostName3)
		require.True(t, res)

		res = authClient.IsApiHostSlackDev(mockHostName4)
		require.True(t, res)
	})
}

func Test_FilterKnownAuthErrors(t *testing.T) {
	tests := map[string]struct {
		err        *slackerror.Error
		filtered   bool
		unfiltered error
	}{
		"returns no error if no error was provided to filter": {
			err:        nil,
			filtered:   false,
			unfiltered: (*slackerror.Error)(nil),
		},
		"filters the already logged out error": {
			err:        slackerror.New(slackerror.ErrAlreadyLoggedOut),
			filtered:   true,
			unfiltered: nil,
		},
		"filters the invalid auth error": {
			err:        slackerror.New(slackerror.ErrInvalidAuth),
			filtered:   true,
			unfiltered: nil,
		},
		"filters the token expired error": {
			err:        slackerror.New(slackerror.ErrTokenExpired),
			filtered:   true,
			unfiltered: nil,
		},
		"filters the token revoked error": {
			err:        slackerror.New(slackerror.ErrTokenRevoked),
			filtered:   true,
			unfiltered: nil,
		},
		"errors if the error is an unexpected error": {
			err:        slackerror.New(slackerror.ErrNotFound),
			filtered:   false,
			unfiltered: slackerror.New(slackerror.ErrNotFound),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			fs := slackdeps.NewFsMock()
			config := config.NewConfig(fs, os)
			io := iostreams.NewIOStreamsMock(config, fs, os)
			io.AddDefaultMocks()
			auth := NewClient(nil, nil, config, io, fs)
			filtered, unfiltered := auth.FilterKnownAuthErrors(ctx, tt.err)
			assert.Equal(t, tt.filtered, filtered)
			assert.Equal(t, tt.unfiltered, unfiltered)
		})
	}
}
