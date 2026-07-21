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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// promptTeamSlackAuthFunc is a package variable so it can be stubbed in tests.
var promptTeamSlackAuthFunc = prompts.PromptTeamSlackAuth

func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	var blocksFlag string
	cmd := &cobra.Command{
		Use:   "preview [flags]",
		Short: "Preview blocks in the Block Kit Builder",
		Long: strings.Join([]string{
			"Open a set of Block Kit blocks in the Block Kit Builder in a web browser.",
			"",
			"Provide blocks with the --blocks flag.",
			"The input is a JSON array of blocks or a JSON object with a \"blocks\" array.",
			"Pass - to --blocks to read from standard input.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Preview blocks passed as a flag value",
				Command: "blocks preview --blocks '[{\"type\":\"divider\"}]'",
			},
			{
				Meaning: "Preview blocks read from a file",
				Command: "blocks preview --blocks - < blocks.json",
			},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewCommandRunE(clients, cmd, blocksFlag, cmd.Flags().Changed("blocks"))
		},
	}
	cmd.Flags().StringVar(&blocksFlag, "blocks", "", "blocks to preview as a JSON array or object\n  (use - to read from standard input)")
	return cmd
}

func previewCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, blocksFlag string, blocksFlagChanged bool) error {
	ctx := cmd.Context()
	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewStart)

	blocksInput, fromStdin, err := resolveBlocksInput(clients, blocksFlag, blocksFlagChanged)
	if err != nil {
		return err
	}

	blocksJSON, err := normalizeBlocksPayload(blocksInput)
	if err != nil {
		return err
	}

	auth, err := selectTeamAuth(ctx, clients, fromStdin)
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

func selectTeamAuth(ctx context.Context, clients *shared.ClientFactory, fromStdin bool) (*types.SlackAuth, error) {
	auths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, err
	}
	if len(auths) == 0 {
		return nil, slackerror.New(slackerror.ErrCredentialsNotFound)
	}
	if fromStdin && clients.Config.TeamFlag == "" && len(auths) > 1 {
		return nil, slackerror.New(slackerror.ErrMissingFlag).
			WithMessage("The team could not be determined").
			WithRemediation("Select a team with the %s flag when reading blocks from standard input", style.Highlight("--team"))
	}
	return promptTeamSlackAuthFunc(ctx, clients, "Select a team to preview blocks for", nil)
}

func resolveBlocksInput(clients *shared.ClientFactory, flagValue string, flagChanged bool) (string, bool, error) {
	switch {
	case flagChanged && flagValue == "-":
		return readStdinBlocks(clients)
	case flagChanged:
		input := strings.TrimSpace(flagValue)
		if input == "" {
			return "", false, missingBlocksError()
		}
		return input, false, nil
	default:
		return "", false, missingBlocksError()
	}
}

func readStdinBlocks(clients *shared.ClientFactory) (string, bool, error) {
	if clients.IO.IsStdinTTY() {
		return "", true, slackerror.New(slackerror.ErrMissingInput).
			WithMessage("No blocks were provided on standard input").
			WithRemediation("Redirect blocks into the command with %s, e.g. %s", style.Highlight("<"), style.Highlight("slack blocks preview --blocks - < blocks.json"))
	}
	piped, err := io.ReadAll(clients.IO.ReadIn())
	if err != nil {
		return "", true, slackerror.Wrap(err, slackerror.ErrMissingInput)
	}
	input := strings.TrimSpace(string(piped))
	if input == "" {
		return "", true, missingBlocksError()
	}
	return input, true, nil
}

func missingBlocksError() error {
	return slackerror.New(slackerror.ErrMissingInput).
		WithMessage("No blocks were provided").
		WithRemediation("Provide blocks with the %s flag, or pass %s to read from standard input", style.Highlight("--blocks"), style.Highlight("--blocks -"))
}

func normalizeBlocksPayload(input string) (string, error) {
	var parsed any
	if err := goutils.JSONUnmarshal([]byte(input), &parsed); err != nil {
		return "", err
	}

	compacted, err := compactJSON(input)
	if err != nil {
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}

	switch value := parsed.(type) {
	case []any:
		return fmt.Sprintf(`{"blocks":%s}`, compacted), nil
	case map[string]any:
		if _, ok := value["blocks"].([]any); !ok {
			return "", slackerror.New(slackerror.ErrInvalidBlocks)
		}
		return compacted, nil
	default:
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}
}

func compactJSON(input string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(input)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func teamOrEnterpriseID(auth *types.SlackAuth) string {
	if auth.IsEnterpriseInstall {
		return auth.EnterpriseID
	}
	return auth.TeamID
}

func buildBlockKitBuilderURL(apiHost string, id string, blocksJSON string) (string, error) {
	parsed, err := url.Parse(apiHost)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrInvalidArguments)
	}
	if parsed.Host == "" {
		return "", slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("The API host %q is not a valid URL", apiHost)
	}
	parsed.Host = "app." + parsed.Host
	parsed.Path = fmt.Sprintf("/block-kit-builder/%s/builder", id)
	parsed.Fragment = blocksJSON
	return parsed.String(), nil
}
