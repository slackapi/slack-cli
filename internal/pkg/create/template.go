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
	"os"
	"strings"

	giturl "github.com/kubescape/go-git-url"
	githubapi "github.com/kubescape/go-git-url/apis/githubapi"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// slackSamplesOrg is the GitHub organization with sample apps and templates
const slackSamplesOrg = "slack-samples"

// trustedTemplateSources lists the GitHub organizations managed by Slack as a
// trusted source of app templates and samples
var trustedTemplateSources = []string{
	slackSamplesOrg,
	"slackapi",
	"slackhq",
}

// Template describes the app template's path and protocol
type Template struct {
	path    string // path can be a local path or remote git URL
	isGit   bool
	isLocal bool
}

// ResolveTemplateURL returns a git-clone compatible URL
func ResolveTemplateURL(templateURL string) (Template, error) {
	template := Template{
		path: strings.TrimSpace(templateURL),
	}

	// Remote git URL
	if strings.HasPrefix(template.path, "ssh://") ||
		strings.HasPrefix(template.path, "git://") ||
		strings.HasPrefix(template.path, "git@") ||
		strings.HasPrefix(template.path, "https://") ||
		strings.HasPrefix(template.path, "http://") {
		template.isGit = true
		return template, nil
	}

	// Local path when it exists
	if _, err := os.Stat(template.path); err == nil {
		template.isLocal = true
		return template, nil
	}

	// Fallback on a relative github.com repo path when formatted as "org/repo" (e.g. 'slackapi/slack-cli')
	if len(strings.Split(template.path, "/")) == 2 {
		template.isGit = true
		template.path = fmt.Sprintf("https://github.com/%s.git", template.path)
		return template, nil
	}

	return Template{}, slackerror.New(slackerror.ErrTemplatePathNotFound).
		WithRemediation(`Check for an existing Slack app template at the path "%s"`, template.path)
}

// GetTemplatePath returns the path a template is located at
//
// Templates on GitHub use an abbreviated format for concise outputs
func (t Template) GetTemplatePath() string {
	if t.isLocal || !t.isGit || strings.HasPrefix(t.path, "ssh://") {
		return t.path
	}
	gitURL, err := giturl.NewGitURL(t.path)
	if err != nil {
		return t.path
	}
	switch gitURL.GetHostName() {
	case githubapi.DEFAULT_HOST:
		return fmt.Sprintf("%s/%s", gitURL.GetOwnerName(), gitURL.GetRepoName())
	default:
		return t.path
	}
}

// IsSample returns if the complete URL points to a sample template
func (t Template) IsSample() bool {
	if t.isLocal || !t.isGit || strings.HasPrefix(t.path, "ssh://") {
		return false
	}
	gitURL, err := giturl.NewGitURL(t.path)
	if err != nil {
		return false
	}
	cloneURL := gitURL.GetHttpCloneURL()
	return strings.HasPrefix(
		cloneURL,
		fmt.Sprintf("https://github.com/%s", slackSamplesOrg),
	)
}

// IsTrusted returns if the template is cloned from a trusted source
func (t Template) IsTrusted() bool {
	if t.isLocal || !t.isGit || strings.HasPrefix(t.path, "ssh://") {
		return false
	}
	gitURL, err := giturl.NewGitURL(t.path)
	if err != nil {
		return false
	}
	cloneURL := gitURL.GetHttpCloneURL()
	for _, org := range trustedTemplateSources {
		if strings.HasPrefix(cloneURL, fmt.Sprintf("https://github.com/%s", org)) {
			return true
		}
	}
	return false
}
