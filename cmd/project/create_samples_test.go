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
	"io"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockGitHubRepos = []create.GithubRepo{
	{
		ID:              1,
		Name:            "deno-mock-repo",
		FullName:        "mock-org/deno-mock-repo",
		Description:     "Just a mock repo",
		Language:        "TypeScript",
		CreatedAt:       "12345",
		StargazersCount: 15,
	},
	{
		ID:              2,
		Name:            "bolt-js-mock-repo",
		FullName:        "mock-org/bolt-js-mock-repo",
		Description:     "Just a mock repo",
		Language:        "JavaScript",
		CreatedAt:       "12345",
		StargazersCount: 10,
	},
	{
		ID:              3,
		Name:            "random-mock-repo",
		FullName:        "mock-org/random-mock-repo",
		Description:     "This is a new sample",
		Language:        "Python",
		CreatedAt:       "12345",
		StargazersCount: 0,
	},
	{
		ID:              4,
		Name:            "deno-mock-repo-2",
		FullName:        "mock-org/deno-mock-repo-2",
		Description:     "This is a popular sample",
		Language:        "TypeScript",
		CreatedAt:       "12345",
		StargazersCount: 50,
	},
}

func TestSamples_PromptSampleSelection(t *testing.T) {
	tests := map[string]struct {
		mockSlackHTTPResponse string
		mockSelection         iostreams.SelectPromptResponse
		expectedRepository    string
		expectedError         error
	}{
		"select sample prompt to create project": {
			mockSlackHTTPResponse: `[
              {
                "name": "deno-announcement-bot",
                "full_name": "slack-samples\/deno-announcement-bot",
                "created_at": "2025-02-05T12:34:56Z",
                "stargazers_count": 100,
                "description": "Preview, post, and manage announcements sent to one or more channels",
                "language": "TypeScript"
              },
              {
                "name": "deno-blank-template",
                "full_name": "slack-samples\/deno-blank-template",
                "created_at": "2025-02-05T12:34:56Z",
                "stargazers_count": 100,
                "description": "A blank template for building modular Slack apps with Deno",
                "language": "Dockerfile"
              }
            ]`,
			mockSelection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "example deno-hello-mock label",
				Index:  1,
			},
			expectedRepository: "slack-samples/deno-blank-template",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			sampler := create.NewMockSampler()
			w := httptest.NewRecorder()
			_, _ = io.WriteString(w, tt.mockSlackHTTPResponse)
			res := w.Result()
			sampler.On("Do", mock.Anything).Return(res, nil)
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On(
				"SelectPrompt",
				mock.Anything,
				"Select a sample to build upon:",
				mock.Anything,
				iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("template"),
				}),
			).Return(tt.mockSelection, nil)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Execute test
			repoName, err := PromptSampleSelection(ctx, clients, sampler)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedRepository, repoName)
		})
	}
}

func TestSamples_FilterRepos(t *testing.T) {
	filteredRepos := filterRepos(mockGitHubRepos, "deno")
	assert.Equal(t, len(filteredRepos), 2, "Expected filteredRepos length to be 2")
}

func TestSamples_SortRepos(t *testing.T) {
	sortedRepos := sortRepos(mockGitHubRepos)
	assert.Equal(t, sortedRepos[0].StargazersCount, 50, "Expected sortedRepos[0].StargazersCount to equal 50")
	assert.Equal(t, sortedRepos[0].Description, "This is a popular sample")
	assert.Equal(t, sortedRepos[3].StargazersCount, 0, "Expected sortedRepos[3].StargazersCount to equal 0")
	assert.Equal(t, sortedRepos[3].Description, "This is a new sample")
}

func TestSamples_CreateSelectOptions(t *testing.T) {
	selectOptions := createSelectOptions(mockGitHubRepos)
	assert.Equal(t, len(selectOptions), 4, "Expected selectOptions length to be 4")
	assert.Contains(t, selectOptions[0], mockGitHubRepos[0].Name, "Expected selectOptions[0] to contain mockGitHubRepos[0].Name")
}
