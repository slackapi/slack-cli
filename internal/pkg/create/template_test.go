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
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplate_ResolveTemplateURL(t *testing.T) {
	tests := map[string]struct {
		url              string
		expectedTemplate Template
		expectedSample   bool
		expectedError    error
	}{
		"supports http:// for a template": {
			url: "http://github.com/slackapi/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "http://github.com/slackapi/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports http:// for a sample": {
			url: "http://github.com/slack-samples/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "http://github.com/slack-samples/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports https:// for a template": {
			url: "https://github.com/slackapi/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "https://github.com/slackapi/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports https:// for a sample": {
			url: "https://github.com/slack-samples/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "https://github.com/slack-samples/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports http://....git for a template": {
			url: "http://github.com/slackapi/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "http://github.com/slackapi/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports http://....git for a sample": {
			url: "http://github.com/slack-samples/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "http://github.com/slack-samples/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports https://....git for a template": {
			url: "https://github.com/slackapi/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "https://github.com/slackapi/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports https://....git for a sample": {
			url: "https://github.com/slack-samples/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "https://github.com/slack-samples/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports ssh:// for a template": {
			url: "ssh://username@host.xz/path/to/repo.git/",
			expectedTemplate: Template{
				path:  "ssh://username@host.xz/path/to/repo.git/",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports ssh:// for a sample": {
			url: "ssh://username@host.xz/slack-samples/repo.git/",
			expectedTemplate: Template{
				path:  "ssh://username@host.xz/slack-samples/repo.git/",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports git:// for a template": {
			url: "git://github.com/slackapi/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "git://github.com/slackapi/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports git:// for a sample": {
			url: "git://github.com/slack-samples/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "git://github.com/slack-samples/bolt-js-getting-started-app",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports git@ for a template": {
			url: "git@github.com:slackapi/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "git@github.com:slackapi/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports git@ for a sample": {
			url: "git@github.com:slack-samples/bolt-js-getting-started-app.git",
			expectedTemplate: Template{
				path:  "git@github.com:slack-samples/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: true,
		},
		"supports github relative url for a template": {
			url: "slackapi/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "https://github.com/slackapi/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports github relative url for a sample": {
			url: "slack-samples/bolt-js-getting-started-app",
			expectedTemplate: Template{
				path:  "https://github.com/slack-samples/bolt-js-getting-started-app.git",
				isGit: true,
			},
			expectedSample: false,
		},
		"supports local relative paths with a separator for a template": {
			url: "../create",
			expectedTemplate: Template{
				path:    "../create",
				isGit:   false,
				isLocal: true,
			},
			expectedSample: false,
		},
		"errors if local path does not exist": {
			url:              filepath.Join("path", "not", "exists"),
			expectedTemplate: Template{},
			expectedError:    slackerror.New(slackerror.ErrTemplatePathNotFound),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			template, err := ResolveTemplateURL(tt.url)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(
					t,
					slackerror.ToSlackError(tt.expectedError).Code,
					slackerror.ToSlackError(err).Code,
				)
			}
			assert.Equal(t, tt.expectedTemplate, template)
		})
	}
}

func TestTemplate_GetTemplatePath(t *testing.T) {
	tests := map[string]struct {
		template     Template
		expectedPath string
	}{
		"returns a local path without change": {
			template: Template{
				isLocal: true,
				path:    "../path/to/template",
			},
			expectedPath: "../path/to/template",
		},
		"returns a github url in shorthand notation": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slack-samples/bolt-starter-app",
			},
			expectedPath: "slack-samples/bolt-starter-app",
		},
		"returns another git url without change": {
			template: Template{
				isGit: true,
				path:  "https://gitlab.com/slack-samples/bolt-starter-app",
			},
			expectedPath: "https://gitlab.com/slack-samples/bolt-starter-app",
		},
		"returns the unchanged path for invalid git urls": {
			template: Template{
				isGit: true,
				path:  ".git",
			},
			expectedPath: ".git",
		},
		"returns the complete ssh path without panic": {
			template: Template{
				isGit: true,
				path:  "ssh://user@example:path/to/sample",
			},
			expectedPath: "ssh://user@example:path/to/sample",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			path := tt.template.GetTemplatePath()
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestTemplate_IsSample(t *testing.T) {
	tests := map[string]struct {
		template       Template
		expectedSample bool
	}{
		"templates from the slack-samples github org are samples": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slack-samples/bolt-starter-app",
			},
			expectedSample: true,
		},
		"templates from the slack-samples github org cloned via git@ are samples": {
			template: Template{
				isGit: true,
				path:  "git@github.com:slack-samples/bolt-starter-app.git",
			},
			expectedSample: true,
		},
		"templates from other github orgs are not samples": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slackapi/bolt-starter-app",
			},
			expectedSample: false,
		},
		"templates matching slack-samples not hosted on github are not samples": {
			template: Template{
				isGit: true,
				path:  "https://gitlab.com/slack-samples/slack-samples",
			},
			expectedSample: false,
		},
		"templates cloned via ssh are not considered samples": {
			template: Template{
				isGit: true,
				path:  "ssh://user@example:path/to/sample",
			},
			expectedSample: false,
		},
		"local paths are not considered samples": {
			template: Template{
				isLocal: true,
				path:    "../path/to/slack-samples/bolt-starter-app",
			},
			expectedSample: false,
		},
		"invalid git urls are not considered samples": {
			template: Template{
				isGit: true,
				path:  ".git",
			},
			expectedSample: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isSample := tt.template.IsSample()
			assert.Equal(t, tt.expectedSample, isSample)
		})
	}
}

func TestTemplate_IsTrusted(t *testing.T) {
	tests := map[string]struct {
		template        Template
		expectedTrusted bool
	}{
		"templates from the slack-samples github org are trusted": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slack-samples/bolt-starter-app",
			},
			expectedTrusted: true,
		},
		"templates from the slack-samples github org cloned via git@ are trusted": {
			template: Template{
				isGit: true,
				path:  "git@github.com:slack-samples/bolt-starter-app.git",
			},
			expectedTrusted: true,
		},
		"templates from the slackapi org are trusted": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slackapi/bolt-starter-app",
			},
			expectedTrusted: true,
		},
		"templates from the slackhq org are trusted": {
			template: Template{
				isGit: true,
				path:  "https://github.com/slackhq/bolt-starter-app",
			},
			expectedTrusted: true,
		},
		"templates from other github orgs are not trusted": {
			template: Template{
				isGit: true,
				path:  "https://github.com/github/bolt-starter-app",
			},
			expectedTrusted: false,
		},
		"templates matching slack-samples not hosted on github are not trusted": {
			template: Template{
				isGit: true,
				path:  "https://gitlab.com/slack-samples/slack-samples",
			},
			expectedTrusted: false,
		},
		"templates cloned via ssh are not considered trusted": {
			template: Template{
				isGit: true,
				path:  "ssh://user@example:path/to/sample",
			},
			expectedTrusted: false,
		},
		"local paths are not considered trusted": {
			template: Template{
				isLocal: true,
				path:    "../path/to/slack-samples/bolt-starter-app",
			},
			expectedTrusted: false,
		},
		"invalid git urls are not considered trusted": {
			template: Template{
				isGit: true,
				path:  ".git",
			},
			expectedTrusted: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isTrusted := tt.template.IsTrusted()
			assert.Equal(t, tt.expectedTrusted, isTrusted)
		})
	}
}
