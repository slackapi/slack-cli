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

package apps

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAppsDelete(t *testing.T) {
	tests := map[string]struct {
		app     types.App
		auth    types.SlackAuth
		unsaved bool
	}{
		"app created on enterprise workspace can be deleted using workspace-level auth": {
			app: types.App{
				AppID:        "A123",
				TeamDomain:   "ws",
				TeamID:       "T123",
				EnterpriseID: "E123",
				IsDev:        true,
			},
			auth: types.SlackAuth{
				EnterpriseID: "E123",
				TeamID:       "E123",
				TeamDomain:   "ws",
				Token:        "xoxb-0123",
			},
		},
		"development app created on enterprise workspace can be deleted using org-level auth": {
			app: types.App{
				AppID:        "A123",
				EnterpriseID: "E123",
				TeamID:       "T123",
				TeamDomain:   "ws",
				IsDev:        true,
			},
			auth: types.SlackAuth{
				EnterpriseID: "E123",
				TeamID:       "E123",
				TeamDomain:   "org",
				Token:        "xoxb-example",
			},
		},
		"app not saved to project files is still deleted without error": {
			app: types.App{
				AppID:      "A001",
				TeamID:     "T002",
				TeamDomain: "ws",
			},
			auth: types.SlackAuth{
				TeamID:     "T002",
				TeamDomain: "ws",
				Token:      "xoxb-001",
			},
			unsaved: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AuthInterface.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api host")
			clientsMock.AuthInterface.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash host")
			clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
				TeamName: &tt.auth.TeamDomain,
				TeamID:   &tt.auth.TeamID,
			}, nil)
			clientsMock.API.On("DeleteApp", mock.Anything, mock.Anything, tt.app.AppID).Return(nil)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			if !tt.unsaved {
				if !tt.app.IsDev {
					err := clients.AppClient().SaveDeployed(ctx, tt.app)
					require.NoError(t, err)
					apps, _, err := clients.AppClient().GetDeployedAll(ctx)
					require.NoError(t, err)
					assert.Equal(t, 1, len(apps))
				} else {
					err := clients.AppClient().SaveLocal(ctx, tt.app)
					require.NoError(t, err)
					apps, err := clients.AppClient().GetLocalAll(ctx)
					require.NoError(t, err)
					assert.Equal(t, 1, len(apps))
				}
			}

			app, err := Delete(ctx, clients, logger.New(nil), tt.app.TeamDomain, tt.app, tt.auth)
			require.NoError(t, err)
			assert.Equal(t, tt.app, app)
			clientsMock.API.AssertCalled(
				t,
				"DeleteApp",
				mock.Anything,
				tt.auth.Token,
				tt.app.AppID,
			)
			if !tt.app.IsDev {
				apps, _, err := clients.AppClient().GetDeployedAll(ctx)
				require.NoError(t, err)
				assert.Equal(t, 0, len(apps))
			} else {
				apps, err := clients.AppClient().GetLocalAll(ctx)
				require.NoError(t, err)
				assert.Equal(t, 0, len(apps))
			}
		})
	}
}
