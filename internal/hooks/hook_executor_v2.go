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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// HookExecutorMessageBoundaryProtocol uses a protocol between the CLI and the SDK where diagnostic info
// and hook responses come in via stdout, and hook responses are wrapped in a string denoting the
// message boundary. Only one message payload can be received.
type HookExecutorMessageBoundaryProtocol struct {
	IO iostreams.IOStreamer
	Fs afero.Fs
}

// generateBoundary is a function for creating boundaries that can be mocked
var generateBoundary = generateRandomBoundary

// Execute processes the data received by the SDK.
func (e *HookExecutorMessageBoundaryProtocol) Execute(ctx context.Context, opts HookExecOpts) (string, error) {
	cmdArgs, cmdArgVars, cmdEnvVars, err := processExecOpts(ctx, opts, e.Fs, e.IO)
	if err != nil {
		return "", err
	}

	if opts.Exec == nil {
		opts.Exec = ShellExec{}
	}

	boundary := generateBoundary()
	cmdArgVars = append(cmdArgVars, "--protocol="+HookProtocolV2.String(), "--boundary="+boundary)

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
	stdout := iostreams.BoundariedWriter{
		Buff:   &buffout,
		Bounds: boundary,
		Stream: iostreams.BufferedWriter{
			Buff: iostreams.FilteredWriter{
				Bounds: boundary,
				Stream: opts.Stdout,
			},
			Stream: e.IO.WriteDebug(ctx),
		},
	}
	stderr := iostreams.BufferedWriter{
		Buff: &bufferr,
		Stream: iostreams.BufferedWriter{
			Buff: iostreams.FilteredWriter{
				Bounds: boundary,
				Stream: opts.Stderr,
			},
			Stream: e.IO.WriteDebug(ctx),
		},
	}

	cmd := opts.Exec.Command(cmdEnvVars, &stdout, stderr, opts.Stdin, cmdArgs[0], cmdArgVars...)
	if err = cmd.Run(); err != nil {
		// Include stderr outputs in error details if these aren't streamed
		details := slackerror.ErrorDetails{}
		if opts.Stderr == nil {
			details = append(details, slackerror.ErrorDetail{Message: strings.TrimSpace(bufferr.String())})
		}
		return "", slackerror.New(slackerror.ErrSDKHookInvocationFailed).
			WithMessage("Error running '%s' command: %s", opts.Hook.Name, err).
			WithDetails(details)
	}
	return buffout.String(), nil
}

// generateRandomBoundary returns the SHA-256 hash of a randomized string for use
// as a message boundary between the CLI and SDK.
//
// Reference: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func generateRandomBoundary() string {
	const alphanumericCharacters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 10

	randomBytes := make([]byte, 0, length)
	for range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericCharacters))))
		if err != nil {
			// Return default value to continue execution
			return "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
		}
		randomBytes = append(randomBytes, alphanumericCharacters[num.Int64()])
	}

	hash := sha256.New()
	hash.Write(randomBytes)
	return hex.EncodeToString(hash.Sum(nil))
}
