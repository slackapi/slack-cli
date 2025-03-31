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

package iostreams

import (
	"context"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// NewIOStreams sets the config and standard I/O for a new IOStreams
func NewIOStreams(config *config.Config, fsFuncs afero.Fs, osFuncs types.Os) *IOStreams {
	var io = &IOStreams{}

	io.config = config
	io.fs = fsFuncs
	io.os = osFuncs

	io.Stdin = os.Stdin
	// Use the stdout and stderr as logger writer
	io.Stdout = log.New(os.Stdout, "", 0)
	io.Stderr = log.New(os.Stderr, "", 0)

	io.exitCode = ExitOK

	return io
}

type ExitCode int32

// Exit codes to return to the shell upon exiting
//
// https://tldp.org/LDP/abs/html/exitcodes.html
const (
	ExitOK     ExitCode = 0
	ExitError  ExitCode = 1
	ExitCancel ExitCode = 130
)

type IOStreams struct {
	config *config.Config
	fs     afero.Fs
	os     types.Os

	Stdin  io.Reader
	Stdout *log.Logger
	Stderr *log.Logger

	exitCode ExitCode
}

type IOStreamer interface {
	// Reader contains implementation of readers that read input streams
	Reader
	// Printer contains implementations of printers that output to display and file
	Printer
	// Writer contains implementations of io.Writer to log and output inputs
	Writer

	// SetCmdIO sets the Cobra command I/O to match the IOStream
	SetCmdIO(cmd *cobra.Command)

	// IsTTY returns true if the device is an interactive terminal
	IsTTY() bool

	// GetExitCode returns the most-recently set desired exit code in a thread safe way
	GetExitCode() ExitCode
	// SetExitCode sets the desired process exit code in a thread safe way
	SetExitCode(code ExitCode)

	// ConfirmPrompt prompts the user for a "yes" or "no" (true or false) value
	// for the message
	ConfirmPrompt(ctx context.Context, message string, defaultValue bool) (bool, error)
	// InputPrompt prompts the user for a string value for the message, which
	// can be required
	InputPrompt(ctx context.Context, message string, cfg InputPromptConfig) (string, error)
	// MultiSelectPrompt prompts the user to choose multiple options
	MultiSelectPrompt(ctx context.Context, message string, options []string) ([]string, error)
	// PasswordPrompt prompts the user for a string value with the message,
	// replacing typed characters with '*'
	PasswordPrompt(ctx context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error)
	// SelectPrompt prompts the user to choose one of the options
	SelectPrompt(ctx context.Context, message string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error)

	// InitLogFile will start the debug info to
	// .slack/logs/slack-debug-[date].log with debug-ability data
	InitLogFile(ctx context.Context) error
	// FinishLogFile will end the debug info to
	// .slack/logs/slack-debug-[date].log with debug-ability data
	FinishLogFile(ctx context.Context)
	// FlushToLogFile flushes messages to the logs and logstash
	FlushToLogFile(ctx context.Context, prefix string, errStr string) error
}

// SetCmdIO sets the Cobra command I/O to match the IOStream
func (io *IOStreams) SetCmdIO(cmd *cobra.Command) {
	cmd.SetIn(io.ReadIn())
	cmd.SetOut(io.WriteOut())
	cmd.SetErr(io.WriteErr())
}

// IsTTY returns true if the device is an interactive terminal
//
// Reference: https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/
func (io *IOStreams) IsTTY() bool {
	if o, err := io.os.Stdout().Stat(); o == nil || err != nil {
		return false
	} else {
		return (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	}
}

// SetExitCode sets the desired exit code in a thread-safe way
func (io *IOStreams) SetExitCode(code ExitCode) {
	atomic.StoreInt32((*int32)(&io.exitCode), int32(code))
}

// GetExitCode returns the most-recently set desired exit code in a thread safe way
func (io *IOStreams) GetExitCode() ExitCode {
	return ExitCode(atomic.LoadInt32((*int32)(&io.exitCode)))
}
