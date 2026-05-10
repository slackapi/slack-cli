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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type cmdFlags struct {
	method  string
	json    string
	data    string
	headers []string
	include bool
}

var flags cmdFlags

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <method> [key=value ...] [flags]",
		Short: "Call any Slack API method",
		Long: strings.Join([]string{
			"Call any Slack API method directly.",
			"",
			"The method argument is the Slack API method name (e.g., \"chat.postMessage\").",
			"Parameters are passed as key=value pairs, a JSON body, or via flags.",
			"",
			"Body format is auto-detected from positional arguments:",
			"  - Multiple key=value args: form-encoded (token in request body)",
			"  - Single arg starting with { or [: JSON (Bearer token in header)",
			"  - No args: token sent in Authorization header",
			"",
			"Use --json to explicitly send a JSON body, or --data for a form-encoded body string.",
			"",
			"Token resolution (in priority order):",
			"  1. --token flag              Explicit token value",
			"  2. --app / --team flags      Install app and use bot token (in project)",
			"  3. SLACK_BOT_TOKEN env var   Bot token (set during slack deploy)",
			"  4. SLACK_USER_TOKEN env var  User token",
			"  5. Interactive prompt        Select from stored workspaces (CLI tooling token)",
			"",
			"See all methods at: https://docs.slack.dev/reference/methods",
			"",
			"Common methods:",
			"  api.test                        Test your API connection",
			"  auth.test                       Check authentication",
			"  chat.postMessage                Send a message to a channel",
			"  chat.update                     Update a message",
			"  chat.delete                     Delete a message",
			"  conversations.list              List channels",
			"  conversations.history           Fetch messages from a channel",
			"  conversations.info              Get channel details",
			"  conversations.members           List members in a channel",
			"  conversations.create            Create a channel",
			"  users.list                      List workspace members",
			"  users.info                      Get user details",
			"  files.upload                    Upload a file",
			"  reactions.add                   Add an emoji reaction",
			"  reactions.list                  List reactions for a user",
			"  bookmarks.add                   Add a bookmark to a channel",
			"  pins.add                        Pin a message",
			"  views.open                      Open a modal view",
			"  views.update                    Update a modal view",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "api auth.test", Meaning: "Test authentication with the current workspace"},
			{Command: "api chat.postMessage channel=C0123456 text=\"Hello\"", Meaning: "Post a message"},
			{Command: "api users.list --team myworkspace", Meaning: "List users in a specific workspace"},
			{Command: `api chat.postMessage --json '{"channel":"C0123456","text":"Hello"}'`, Meaning: "Send a JSON body"},
			{Command: "api auth.test --include", Meaning: "Show HTTP status and response headers"},
			{Command: "api conversations.history -X GET channel=C0123456", Meaning: "Use GET method"},
		}),
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPICommand(cmd, clients, args)
		},
	}

	cmd.Flags().StringVarP(&flags.method, "method", "X", "POST", "HTTP method for the request")
	cmd.Flags().StringVar(&flags.json, "json", "", "JSON request body (uses Bearer token in Authorization header)")
	cmd.Flags().StringVar(&flags.data, "data", "", "form-encoded request body string (e.g. \"key1=val1&key2=val2\")")
	cmd.Flags().StringSliceVarP(&flags.headers, "header", "H", nil, "additional HTTP headers (format: \"Key: Value\")")
	cmd.Flags().BoolVarP(&flags.include, "include", "i", false, "include HTTP status code and response headers in output")
	cmd.MarkFlagsMutuallyExclusive("json", "data")

	return cmd
}

func runAPICommand(cmd *cobra.Command, clients *shared.ClientFactory, args []string) error {
	ctx := cmd.Context()
	method := args[0]
	params := args[1:]

	token, err := resolveToken(ctx, clients)
	if err != nil {
		return err
	}

	apiHost := clients.Config.APIHostResolved
	if apiHost == "" {
		apiHost = "https://slack.com"
	}
	apiClient := api.NewClient(nil, apiHost, clients.IO)

	var bodyReader *strings.Reader
	var contentType string

	switch {
	case flags.json != "":
		contentType = "application/json; charset=utf-8"
		bodyReader = strings.NewReader(flags.json)
	case flags.data != "":
		contentType = "application/x-www-form-urlencoded"
		formData := flags.data
		if !strings.Contains(formData, "token=") {
			if formData != "" {
				formData = formData + "&token=" + url.QueryEscape(token)
			} else {
				formData = "token=" + url.QueryEscape(token)
			}
		}
		bodyReader = strings.NewReader(formData)
		token = ""
	case len(params) == 1 && (strings.HasPrefix(params[0], "{") || strings.HasPrefix(params[0], "[")):
		contentType = "application/json; charset=utf-8"
		bodyReader = strings.NewReader(params[0])
	case len(params) > 0:
		contentType = "application/x-www-form-urlencoded"
		values := url.Values{}
		values.Set("token", token)
		for _, param := range params {
			key, value, ok := strings.Cut(param, "=")
			if !ok {
				return slackerror.New(slackerror.ErrInvalidArguments).
					WithMessage("invalid parameter %q: must be in key=value format", param)
			}
			values.Set(key, value)
		}
		bodyReader = strings.NewReader(values.Encode())
		token = ""
	default:
		contentType = "application/x-www-form-urlencoded"
		values := url.Values{}
		values.Set("token", token)
		bodyReader = strings.NewReader(values.Encode())
		token = ""
	}

	customHeaders := map[string]string{}
	for _, h := range flags.headers {
		key, value, ok := strings.Cut(h, ":")
		if !ok {
			return slackerror.New(slackerror.ErrInvalidArguments).
				WithMessage("invalid header %q: must be in \"Key: Value\" format", h)
		}
		customHeaders[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	resp, err := apiClient.RawRequest(ctx, flags.method, method, token, bodyReader, contentType, customHeaders)
	if err != nil {
		return err
	}

	if flags.include {
		fmt.Fprintf(cmd.OutOrStdout(), "HTTP %d\n", resp.StatusCode)
		for key, values := range resp.Header {
			for _, v := range values {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", key, v)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	output := resp.Body
	// Pretty-print for interactive terminals, compact for piped output (gh/git convention)
	if clients.IO.IsTTY() {
		var indented bytes.Buffer
		if json.Indent(&indented, resp.Body, "", "    ") == nil {
			output = indented.Bytes()
		}
	}
	fmt.Fprint(cmd.OutOrStdout(), string(output))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return slackerror.New("api_request_failed").
			WithMessage("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

func resolveToken(ctx context.Context, clients *shared.ClientFactory) (string, error) {
	if clients.Config.TokenFlag != "" {
		return clients.Config.TokenFlag, nil
	}

	if sdkConfigExists, _ := clients.SDKConfig.Exists(); sdkConfigExists {
		selected, err := prompts.AppSelectPrompt(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
		if err == nil && selected.App.AppID != "" {
			token, err := installAndGetBotToken(ctx, clients, selected)
			if err == nil && token != "" {
				return token, nil
			}
		}
	}

	if token := os.Getenv("SLACK_BOT_TOKEN"); token != "" {
		return token, nil
	}

	if token := os.Getenv("SLACK_USER_TOKEN"); token != "" {
		return token, nil
	}

	clients.IO.PrintDebug(ctx, "Using CLI tooling token which has limited API scopes. Set SLACK_BOT_TOKEN or use --token for full access.")
	auth, err := prompts.PromptTeamSlackAuth(ctx, clients, "Select a workspace")
	if err != nil {
		return "", err
	}
	return auth.Token, nil
}

func installAndGetBotToken(ctx context.Context, clients *shared.ClientFactory, selected prompts.SelectedApp) (string, error) {
	manifestSource, _ := clients.Config.ProjectConfig.GetManifestSource(ctx)
	var slackManifest types.SlackYaml
	var err error
	if manifestSource.Equals(config.ManifestSourceRemote) {
		slackManifest, err = clients.AppClient().Manifest.GetManifestRemote(ctx, selected.Auth.Token, selected.App.AppID)
	} else {
		slackManifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
	}
	if err != nil {
		return "", err
	}

	manifest := slackManifest.AppManifest
	botScopes := []string{}
	if manifest.OAuthConfig != nil && manifest.OAuthConfig.Scopes != nil {
		botScopes = manifest.OAuthConfig.Scopes.Bot
	}
	outgoingDomains := []string{}
	if manifest.OutgoingDomains != nil {
		outgoingDomains = *manifest.OutgoingDomains
	}

	result, _, err := clients.API().DeveloperAppInstall(ctx, clients.IO, selected.Auth.Token, selected.App, botScopes, outgoingDomains, "", false)
	if err != nil {
		return "", err
	}

	return result.APIAccessTokens.Bot, nil
}
