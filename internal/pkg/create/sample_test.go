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

package create

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSamples_GetSampleRepos(t *testing.T) {
	tests := map[string]struct {
		slackAPIResponse string
		githubResponse   string
		expectedResults  []GithubRepo
		expectedError    *slackerror.Error
	}{
		"returns samples repos from slack api when available": {
			slackAPIResponse: `[{
                "name": "bolt-starter-app",
                "full_name": "slack-samples/bolt-starter-app",
                "created_at": "2025-02-11T12:34:56Z",
                "stargazers_count": 4,
                "description": "a low level messaging bot",
                "language": "c"
            }]`,
			expectedResults: []GithubRepo{
				{
					Name:            "bolt-starter-app",
					FullName:        "slack-samples/bolt-starter-app",
					CreatedAt:       "2025-02-11T12:34:56Z",
					StargazersCount: 4,
					Description:     "a low level messaging bot",
					Language:        "c",
				},
			},
		},
		"returns samples repos from github as fallback": {
			slackAPIResponse: `[{
                "name": "bolt-broken-wifi",
                "full-na`,
			githubResponse: `[{
                "name": "bolt-example-app",
                "full_name": "slack-samples/bolt-example-app",
                "created_at": "2025-02-12T20:00:00Z",
                "stargazers_count": 12,
                "description": "a collection of sound",
                "language": "music"
            }]`,
			expectedResults: []GithubRepo{
				{
					Name:            "bolt-example-app",
					FullName:        "slack-samples/bolt-example-app",
					CreatedAt:       "2025-02-12T20:00:00Z",
					StargazersCount: 12,
					Description:     "a collection of sound",
					Language:        "music",
				},
			},
		},
		"errors if both requests for the latest samples fail": {
			slackAPIResponse: "[",
			githubResponse:   "[{",
			expectedError:    slackerror.New(slackerror.ErrSampleCreate),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientMock := NewMockSampler()
			w := httptest.NewRecorder()
			_, _ = io.WriteString(w, fmt.Sprint(tt.slackAPIResponse))
			res := w.Result()
			req, err := http.NewRequest("GET", "https://www.slack.com/slack-samples/repositories", nil)
			require.NoError(t, err)
			req.Header.Set("Accept", "application/json")
			clientMock.On("Do", req).Return(res, nil)
			w = httptest.NewRecorder()
			_, _ = io.WriteString(w, fmt.Sprint(tt.githubResponse))
			res = w.Result()
			req, err = http.NewRequest("GET", "https://api.github.com/orgs/slack-samples/repos?type=public", nil)
			require.NoError(t, err)
			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
			clientMock.On("Do", req).Return(res, nil)

			result, err := GetSampleRepos(clientMock)
			if tt.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(err).Code, tt.expectedError.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResults, result)
			}
		})
	}
}
