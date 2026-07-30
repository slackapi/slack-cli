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
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	var blocksFlag string
	cmd := &cobra.Command{
		Use:   "preview [flags]",
		Short: "Preview blocks in Block Kit Builder",
		Long: strings.Join([]string{
			"Preview a set of Block Kit blocks with Block Kit Builder in a web browser.",
			"",
			"Provide blocks with the --blocks flag.",
			"The input is a JSON array of blocks or a JSON object with a \"blocks\" array.",
			"Pass - to --blocks, or omit all flags, to read from standard input.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Preview blocks passed as a flag value",
				Command: "blocks preview --blocks '[{\"type\":\"divider\"}]'",
			},
			{
				Meaning: "Preview blocks read from a file",
				Command: "blocks preview < blocks.json",
			},
			{
				Meaning: "Preview blocks read from a redirect and scoped to a team",
				Command: "blocks preview --team T0123456 --blocks - < blocks.json",
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

	blocksInput, err := resolveBlocksInput(clients, blocksFlag, blocksFlagChanged)
	if err != nil {
		return err
	}

	blocksJSON, err := normalizeBlocksPayload(blocksInput)
	if err != nil {
		return err
	}

	auth, err := resolveTeamAuth(ctx, clients)
	if err != nil {
		return err
	}

	builderURL, err := buildBlockKitBuilderURL(clients.API().Host(), teamOrEnterpriseID(auth), blocksJSON)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "bricks",
		Text:  "Blocks Preview",
		Secondary: []string{
			builderURL,
		},
	}))
	clients.Browser().OpenURL(builderURL)

	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewSuccess, builderURL)
	return nil
}

func resolveTeamAuth(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	if clients.Config.TeamFlag == "" {
		return nil, nil
	}
	auths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, err
	}
	for i, auth := range auths {
		if clients.Config.TeamFlag == auth.TeamID || clients.Config.TeamFlag == auth.TeamDomain {
			return &auths[i], nil
		}
	}
	return nil, slackerror.New(slackerror.ErrTeamNotFound)
}

func resolveBlocksInput(clients *shared.ClientFactory, flagValue string, flagChanged bool) (string, error) {
	switch {
	case flagChanged && flagValue == "-":
		if clients.IO.IsStdinTTY() {
			return "", slackerror.New(slackerror.ErrMissingInput).
				WithMessage("No blocks were provided on standard input").
				WithRemediation("Redirect blocks into the command with \"<\", e.g. %s", style.Commandf("blocks preview < blocks.json", false))
		}
		return readStdinBlocks(clients)
	case flagChanged:
		input := strings.TrimSpace(flagValue)
		if input == "" {
			return "", missingBlocksError()
		}
		return input, nil
	case !clients.IO.IsStdinTTY():
		return readStdinBlocks(clients)
	default:
		return "", missingBlocksError()
	}
}

func readStdinBlocks(clients *shared.ClientFactory) (string, error) {
	input, err := clients.IO.ReadInAll()
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrMissingInput)
	}
	if input == "" {
		return "", missingBlocksError()
	}
	return input, nil
}

func missingBlocksError() error {
	return slackerror.New(slackerror.ErrMissingInput).
		WithMessage("No blocks were provided").
		WithRemediation("Provide blocks with the %s flag, or pass %s to read from standard input", style.Highlight("--blocks"), style.Highlight("--blocks -"))
}

func normalizeBlocksPayload(input string) (string, error) {
	compacted, err := goutils.CompactJSON([]byte(input))
	if err != nil {
		return "", slackerror.JSONUnmarshalError(err, []byte(input))
	}
	trimmed := bytes.TrimSpace(compacted)
	if len(trimmed) == 0 {
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}

	switch trimmed[0] {
	case '[':
		return fmt.Sprintf(`{"blocks":%s}`, trimmed), nil
	case '{':
		var object map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &object); err != nil {
			return "", slackerror.New(slackerror.ErrInvalidBlocks)
		}
		blocks, ok := object["blocks"]
		if !ok || len(bytes.TrimSpace(blocks)) == 0 || bytes.TrimSpace(blocks)[0] != '[' {
			return "", slackerror.New(slackerror.ErrInvalidBlocks)
		}
		return string(trimmed), nil
	default:
		return "", slackerror.New(slackerror.ErrInvalidBlocks)
	}
}

func teamOrEnterpriseID(auth *types.SlackAuth) string {
	if auth == nil {
		return ""
	}
	if auth.IsEnterpriseInstall {
		return auth.EnterpriseID
	}
	return auth.TeamID
}

func buildBlockKitBuilderURL(apiHost string, teamOrEnterpriseID string, blocksJSON string) (string, error) {
	parsed, err := url.Parse(apiHost)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrInvalidArguments)
	}
	if parsed.Host == "" {
		return "", slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("The API host %q is not a valid URL", apiHost)
	}
	parsed.Host = "app." + parsed.Host
	if teamOrEnterpriseID != "" {
		parsed.Path = fmt.Sprintf("/block-kit-builder/%s/builder", teamOrEnterpriseID)
	} else {
		parsed.Path = "/block-kit-builder"
	}
	parsed.Fragment = blocksJSON
	return parsed.String(), nil
}
