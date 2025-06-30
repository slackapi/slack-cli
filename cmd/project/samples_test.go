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
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	createPkg "github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSamplesCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"creates a template from a trusted sample": {
			CmdArgs: []string{"my-sample-app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createPkg.GetSampleRepos = func(client createPkg.Sampler) ([]createPkg.GithubRepo, error) {
					repos := []createPkg.GithubRepo{
						{
							Name:            "deno-starter-template",
							FullName:        "slack-samples/deno-starter-template",
							CreatedAt:       "2025-02-11T12:34:56Z",
							StargazersCount: 4,
							Description:     "a mock starter template for deno",
							Language:        "deno",
						},
					}
					return repos, nil
				}
				cm.IO.On("SelectPrompt", mock.Anything, "Select a language:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Index:  2,
							Prompt: true,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a sample to build upon:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Index:  0,
							Prompt: true,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an app:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Option: "slack-samples/deno-starter-template",
							Flag:   true,
						},
						nil,
					)
				CreateFunc = func(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, createArgs createPkg.CreateArgs) (appDirPath string, err error) {
					return createArgs.AppName, nil
				}
			},
			ExpectedOutputs: []string{
				"cd my-sample-app/",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				for _, call := range cm.IO.Calls {
					switch call.Method {
					case "SelectPrompt":
						args := call.Arguments
						opts := args.Get(3).(iostreams.SelectPromptConfig)
						flag := opts.Flag
						switch args.String(1) {
						case "Select a sample to build upon:":
							require.Equal(t, "template", flag.Name)
							assert.Equal(t, "", flag.Value.String())
						case "Select a template to build from:":
							require.Equal(t, "template", flag.Name)
							assert.Equal(t, "slack-samples/deno-starter-template", flag.Value.String())
						}
					}
				}
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewSamplesCommand(cf)
	})
}
