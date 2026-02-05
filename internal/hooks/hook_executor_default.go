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

package hooks

import (
	"bytes"
	"context"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// HookExecutorDefaultProtocol uses the original protocol between the CLI and the SDK where diagnostic info
// and hook responses come in via stdout. Data outside the expected JSON payload is ignored, with the
// exception of the 'start' hook, for which it is printed.
type HookExecutorDefaultProtocol struct {
	IO iostreams.IOStreamer
}

// Execute processes the data received by the SDK.
func (e *HookExecutorDefaultProtocol) Execute(ctx context.Context, opts HookExecOpts) (string, error) {
	cmdArgs, cmdArgVars, cmdEnvVars, err := processExecOpts(opts)
	if err != nil {
		return "", err
	}

	if opts.Exec == nil {
		opts.Exec = ShellExec{}
	}

	e.IO.PrintDebug(ctx,
		"starting hook command: %s %s\n", cmdArgs[0], strings.Join(cmdArgVars, " "),
	)
	defer func() {
		e.IO.PrintDebug(ctx,
			"finished hook command: %s %s\n", cmdArgs[0], strings.Join(cmdArgVars, " "),
		)
	}()

	buffout := bytes.Buffer{}
	bufferr := bytes.Buffer{}
	stdout := iostreams.BufferedWriter{
		Buff: &buffout,
		Stream: iostreams.BufferedWriter{
			Buff:   opts.Stdout,
			Stream: e.IO.WriteDebug(ctx),
		},
	}
	stderr := iostreams.BufferedWriter{
		Buff: &bufferr,
		Stream: iostreams.BufferedWriter{
			Buff:   opts.Stderr,
			Stream: e.IO.WriteDebug(ctx),
		},
	}

	cmd := opts.Exec.Command(cmdEnvVars, stdout, stderr, opts.Stdin, cmdArgs[0], cmdArgVars...)
	err = cmd.Run()

	response := strings.TrimSpace(buffout.String())
	if err != nil {
		// Include stderr outputs in error details if these aren't streamed
		details := slackerror.ErrorDetails{}
		if opts.Stderr == nil {
			details = append(details, slackerror.ErrorDetail{Message: strings.TrimSpace(bufferr.String())})
		}
		return "", slackerror.New(slackerror.ErrSDKHookInvocationFailed).
			WithMessage("Error running '%s' command: %s", opts.Hook.Name, err).
			WithDetails(details)
	}

	// Special handling for the baseline protocol for the `start` hook
	if opts.Hook.Name == "Start" {
		// All output except for the last line can be displayed to the user.
		// The last line should contain stringified JSON of the result object sent as a response.
		lines := strings.Split(string(response), "\n")
		response = lines[len(lines)-1]
		excludesLastLine := lines[0 : len(lines)-1]
		_, _ = e.IO.WriteOut().Write([]byte(strings.Join(excludesLastLine, "\n") + "\n"))
	}

	return response, nil
}
