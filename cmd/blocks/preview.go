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

package blocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// promptTeamSlackAuthFunc selects the team whose Block Kit Builder should open.
// It is a package variable so that it can be stubbed in tests.
var promptTeamSlackAuthFunc = prompts.PromptTeamSlackAuth

func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview [blocks]",
		Short: "Preview blocks in the Block Kit Builder",
		Long: strings.Join([]string{
			"Open a set of Block Kit blocks in the Block Kit Builder in a web browser.",
			"",
			"Blocks can be passed as an argument or piped in through standard input. The",
			"input is a JSON array of blocks or a JSON object with a \"blocks\" array.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Preview blocks passed as an argument",
				Command: "blocks preview '[{\"type\":\"divider\"}]'",
			},
			{
				Meaning: "Preview blocks piped from a file",
				Command: "blocks preview < blocks.json",
			},
		}),
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return previewCommandPreRunE(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewCommandRunE(clients, cmd, args)
		},
	}
	return cmd
}

// previewCommandPreRunE gates the command behind the block-kit-builder experiment
func previewCommandPreRunE(clients *shared.ClientFactory) error {
	if !clients.Config.WithExperimentOn(experiment.BlockKitBuilder) {
		return slackerror.New(slackerror.ErrMissingExperiment).
			WithMessage("The blocks preview command is experimental").
			WithRemediation("Enable this command with the %s flag", style.Highlight("--experiment "+string(experiment.BlockKitBuilder)))
	}
	return nil
}

// previewCommandRunE reads blocks from an argument or standard input and opens
// them in the Block Kit Builder for the selected team
func previewCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewStart)

	blocksInput, err := readBlocksInput(clients, args)
	if err != nil {
		return err
	}

	blocksJSON, err := normalizeBlocksPayload(blocksInput)
	if err != nil {
		return err
	}

	auth, err := promptTeamSlackAuthFunc(ctx, clients, "Select a team to preview blocks for", nil)
	if err != nil {
		return err
	}

	builderURL, err := buildBlockKitBuilderURL(clients.API().Host(), teamOrEnterpriseID(auth), blocksJSON)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "eyes",
		Text:  "Block Kit Builder",
		Secondary: []string{
			builderURL,
		},
	}))
	clients.Browser().OpenURL(builderURL)

	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewSuccess, builderURL)
	return nil
}

// readBlocksInput returns the blocks provided as an argument or, when no
// argument is given, read from standard input
func readBlocksInput(clients *shared.ClientFactory, args []string) (string, error) {
	if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}

	piped, err := io.ReadAll(clients.IO.ReadIn())
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrMissingInput)
	}
	input := strings.TrimSpace(string(piped))
	if input == "" {
		return "", slackerror.New(slackerror.ErrMissingInput).
			WithMessage("No blocks were provided").
			WithRemediation("Provide blocks as an argument or pipe them through standard input")
	}
	return input, nil
}

// normalizeBlocksPayload parses the provided input and returns a compact JSON
// string in the shape expected by the Block Kit Builder: {"blocks":[...]}.
// The input may be a bare array of blocks or an object containing a "blocks" key.
func normalizeBlocksPayload(input string) (string, error) {
	var parsed any
	if err := goutils.JSONUnmarshal([]byte(input), &parsed); err != nil {
		return "", err
	}

	var payload map[string]any
	switch value := parsed.(type) {
	case []any:
		payload = map[string]any{"blocks": value}
	case map[string]any:
		if _, ok := value["blocks"].([]any); !ok {
			return "", slackerror.New(slackerror.ErrInvalidBlocks)
		}
		payload = value
	default:
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}

	compact, err := json.Marshal(payload)
	if err != nil {
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}
	return string(compact), nil
}

// teamOrEnterpriseID returns the enterprise ID for enterprise installs and the
// team ID otherwise, matching the identifier used in Block Kit Builder URLs
func teamOrEnterpriseID(auth *types.SlackAuth) string {
	if auth.EnterpriseID != "" {
		return auth.EnterpriseID
	}
	return auth.TeamID
}

// buildBlockKitBuilderURL constructs the Block Kit Builder URL for the given
// API host, team or enterprise ID, and compact blocks JSON. The blocks JSON is
// placed in the URL fragment, which url.URL.String percent-encodes.
func buildBlockKitBuilderURL(apiHost string, id string, blocksJSON string) (string, error) {
	parsed, err := url.Parse(apiHost)
	if err != nil {
		return "", err
	}
	parsed.Host = "app." + parsed.Host
	parsed.Path = fmt.Sprintf("/block-kit-builder/%s/builder", id)
	parsed.Fragment = blocksJSON
	return parsed.String(), nil
}
