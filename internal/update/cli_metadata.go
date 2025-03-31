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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CLIReleaseInfo stores information about most recent release's version and release date
type CLIReleaseInfo struct {
	SlackCLI struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Releases    []struct {
			Version     string `json:"version"`
			ReleaseDate string `json:"release_date"`
		} `json:"releases"`
	} `json:"slack-cli"`
}

type LatestCLIRelease struct {
	Version string
}

type Metadata struct {
	httpClient HTTPClient
}

// checkForUpdate returns the release info from CLI meta including the most recent CLI version that is an upgrade to the currentVersion parameter or nil if no upgrade is available.
func (md *Metadata) CheckForUpdate(ctx context.Context, url, currentVersion string) (*LatestCLIRelease, error) {
	releaseInfo, err := md.latestCLIReleaseInfo(url)
	if err != nil {
		return nil, err
	}

	updateAvailable, err := SemVerGreaterThan(releaseInfo.Version, currentVersion)
	if err != nil {
		return nil, err
	}
	if updateAvailable {
		return releaseInfo, nil
	}
	return nil, nil
}

// latestCLIReleaseInfo return CLIReleaseInfo that describes the latest release version from CLI metadata endpoint.
func (md *Metadata) latestCLIReleaseInfo(url string) (*LatestCLIRelease, error) {
	var cliRelease CLIReleaseInfo

	if err := md.httpClientRequest("GET", url, &cliRelease); err != nil {
		return nil, err
	}

	latestCLIRelease := LatestCLIRelease{
		Version: cliRelease.SlackCLI.Releases[0].Version,
	}

	return &latestCLIRelease, nil
}

// httpClientRequest will make an HTTP request and unmarshal the response to data.
func (md *Metadata) httpClientRequest(method string, host string, data interface{}) error {
	if md.httpClient == nil {
		md.httpClient = &http.Client{}
	}

	req, err := http.NewRequest(method, host, nil)
	if err != nil {
		return err
	}

	res, err := md.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.WithStack(fmt.Errorf("Slack CLI metadata responded with an unexpected status code %d from url %s", res.StatusCode, host))
	}

	var bytes []byte
	bytes, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}

	return nil
}
