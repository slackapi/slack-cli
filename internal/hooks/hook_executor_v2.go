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

package hooks

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// HookExecutorMessageBoundaryProtocol uses a protocol between the CLI and the SDK where diagnostic info
// and hook responses come in via stdout, and hook responses are wrapped in a string denoting the
// message boundary. Only one message payload can be received.
type HookExecutorMessageBoundaryProtocol struct {
	IO iostreams.IOStreamer
}

// generateBoundary is a function for creating boundaries that can be mocked
var generateBoundary = generateMD5FromRandomString

// Execute processes the data received by the SDK.
func (e *HookExecutorMessageBoundaryProtocol) Execute(ctx context.Context, opts HookExecOpts) (string, error) {
	cmdArgs, cmdArgVars, cmdEnvVars, err := processExecOpts(opts)
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
		return "", slackerror.New(slackerror.ErrSDKHookInvocationFailed).
			WithMessage("Error running '%s' command: %s", opts.Hook.Name, err)
	}
	return buffout.String(), nil
}

// generateMD5FromRandomString returns the MD5 hash of a randomized string.
//
// Reference: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func generateMD5FromRandomString() string {
	const alphanumericCharacters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 10

	randomBytes := make([]byte, 0)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericCharacters))))
		if err != nil {
			return "3561f3a3c5576e2ce0dc0d1e268bb9b2" // Return default value to continue execution
		}
		randomBytes = append(randomBytes, alphanumericCharacters[num.Int64()])
	}

	MD5Hash := md5.New()
	s := MD5Hash.Sum(randomBytes)

	return hex.EncodeToString(s)
}
