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

package update

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// HTTPClientMock is a mock for the Slack CLI Metadata HTTPClient
type HTTPClientMock struct {
	mock.Mock
}

// Do is a mock that tracks the calls and returns the mocked http.Response and error
func (m *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called()
	return args.Get(0).(*http.Response), args.Error(1)
}

// Test_CLI_Metadata_CheckForUpdate tests different responses from Slack CLI metadata.
func Test_CLI_Metadata_CheckForUpdate(t *testing.T) {
	const metaDataURL = "https://api.slack.com/slackcli/metadata.json"

	scenarios := map[string]struct {
		CurrentVersion string
		LatestVersion  string
		ExpectsResult  bool
	}{
		"latest is an upgrade": {
			CurrentVersion: "v0.0.1",
			LatestVersion:  "v1.0.0",
			ExpectsResult:  true,
		},
		"latest is newer than version build from source (latest is an upgrade)": {
			CurrentVersion: "v1.2.3-123-gdeadbeef",
			LatestVersion:  "v1.2.4",
			ExpectsResult:  true,
		},
		"latest is current (latest not upgrade)": {
			CurrentVersion: "v1.0.0",
			LatestVersion:  "v1.0.0",
			ExpectsResult:  false,
		},
		"latest is older (latest not upgrade)": {
			CurrentVersion: "v0.10.0-pre.1",
			LatestVersion:  "v0.9.0",
			ExpectsResult:  false,
		},
	}

	for name, s := range scenarios {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// Mock an http.Response for the GitHub API
			w := httptest.NewRecorder()
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{ "slack-cli": { "releases": [ { "version": "%s" } ] } }`,
				s.LatestVersion,
			))
			res := w.Result()

			// Set up httpClientMock to use mock http.Response
			httpClientMock := new(HTTPClientMock)
			httpClientMock.On("Do").Return(res, nil)

			// Check for an update
			md := Metadata{httpClient: httpClientMock}
			releaseInfo, err := md.CheckForUpdate(ctx, metaDataURL, s.CurrentVersion)

			// Assert expected results
			if s.ExpectsResult {
				require.Equal(t, s.LatestVersion, releaseInfo.Version)
				require.Nil(t, err)
			} else {
				require.Nil(t, releaseInfo)
				require.Nil(t, err)
			}
		})
	}
}
