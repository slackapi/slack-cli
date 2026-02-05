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

package create

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Sampler requests samples from various URLs
type Sampler interface {
	Do(req *http.Request) (res *http.Response, err error)
}

// GithubRepo is the single repository data structure
type GithubRepo struct {
	ID              int
	Name            string
	FullName        string `json:"full_name"`
	Description     string
	Language        string
	CreatedAt       string `json:"created_at"`
	StargazersCount int    `json:"stargazers_count"`
}

// GetSampleRepos requests a list of samples from Slack API and falls back to
// the GitHub API
var GetSampleRepos = func(client Sampler) ([]GithubRepo, error) {
	repos, err := getSampleReposFromSlack(client)
	if err != nil || len(repos) == 0 {
		repos, err = getSampleReposFromGitHub(client)
		if err != nil {
			return []GithubRepo{}, slackerror.Wrap(err, slackerror.ErrSampleCreate)
		}
	}
	return repos, nil
}

// getSampleReposFromSlack makes a call to the Slack.com and retrieves
// the repositories available in the Slack Samples Organization,
// unmarshalling the response into a GithubRepo struct
func getSampleReposFromSlack(client Sampler) ([]GithubRepo, error) {
	req, err := http.NewRequest("GET", "https://www.slack.com/slack-samples/repositories", nil)
	if err != nil {
		return []GithubRepo{}, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return []GithubRepo{}, err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%v (Slack.com)", res.Status)
	}
	if err != nil {
		return nil, err
	}

	sampleRepos := []GithubRepo{}
	err = json.Unmarshal(body, &sampleRepos)
	if err != nil {
		return nil, err
	}
	return sampleRepos, nil
}

// getSampleReposFromGitHub makes a call to the GitHub API and retrieves
// the repositories available in the Slack Samples Organization,
// unmarshalling the response into a GithubRepo struct
func getSampleReposFromGitHub(client Sampler) ([]GithubRepo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/orgs/slack-samples/repos?type=public", nil)
	if err != nil {
		return []GithubRepo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := client.Do(req)
	if err != nil {
		return []GithubRepo{}, err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%v (GitHub API)", res.Status)
	}
	if err != nil {
		return nil, err
	}

	sampleRepos := []GithubRepo{}
	err = json.Unmarshal(body, &sampleRepos)
	if err != nil {
		return nil, err
	}
	return sampleRepos, nil
}
