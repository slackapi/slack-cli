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
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync/atomic"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

type IOStreamsMock struct {
	mock.Mock

	config *config.Config
	fs     *slackdeps.FsMock
	os     *slackdeps.OsMock

	Stdin  io.Reader
	Stdout *log.Logger
	stdout *bytes.Buffer
	Stderr *log.Logger

	exitCode ExitCode
}

// NewIOStreamsMock creates a new IOStream with buffers that can be read for tests.
func NewIOStreamsMock(config *config.Config, fsm *slackdeps.FsMock, osm *slackdeps.OsMock) *IOStreamsMock {
	m := &IOStreamsMock{}

	m.config = config
	m.fs = fsm
	m.os = osm

	// Add buffers as writers/reader
	stdin := &bytes.Buffer{}
	m.Stdin = stdin

	m.stdout = &bytes.Buffer{}
	m.Stdout = &log.Logger{}
	m.Stdout.SetOutput(m.stdout)

	stderr := &bytes.Buffer{}
	m.Stderr = &log.Logger{}
	m.Stderr.SetOutput(stderr)

	m.exitCode = ExitOK

	return m
}

// AddDefaultMocks prepares default mock methods to fallback to
func (m *IOStreamsMock) AddDefaultMocks() {
	m.On("IsTTY").Return(false)
	m.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	m.On("PrintTrace", mock.Anything, mock.Anything, mock.MatchedBy(func(args []string) bool { return true }))
	m.On("PrintWarning", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	m.On("WriteDebug", mock.Anything)
	m.On("WriteSecondary", mock.Anything).Unset()
	m.On("WriteSecondary", m.Stdout.Writer()).Return(WriteSecondarier{Writer: m.Stdout.Writer()})
	m.On("WriteSecondary", m.Stderr.Writer()).Return(WriteSecondarier{Writer: m.Stderr.Writer()})
}

// DumpLogs can sometimes be useful when debugging unit tests
func (m *IOStreamsMock) DumpLogs() {
	fmt.Println(m.stdout)
}

// SetCmdIO sets the Cobra command I/O to be the same as the IOStreamsMock
func (m *IOStreamsMock) SetCmdIO(cmd *cobra.Command) {
	cmd.SetIn(m.ReadIn())
	cmd.SetOut(m.WriteOut())
	cmd.SetErr(m.WriteErr())
}

func (m *IOStreamsMock) IsTTY() bool {
	args := m.Called()
	return args.Bool(0)
}

// SetExitCode sets the desired exit code in a thread-safe way
func (io *IOStreamsMock) SetExitCode(code ExitCode) {
	atomic.StoreInt32((*int32)(&io.exitCode), int32(code))
}

// GetExitCode returns the most-recently set desired exit code in a thread safe way
func (io *IOStreamsMock) GetExitCode() ExitCode {
	return ExitCode(atomic.LoadInt32((*int32)(&io.exitCode)))
}

// InitLogFile mocks starting the debug info to
// .slack/logs/slack-debug-[date].log with debug-ability data
func (m *IOStreamsMock) InitLogFile(ctx context.Context) error {
	return nil
}

// FinishLogFile mocks ending the debug info to
// .slack/logs/slack-debug-[date].log with debug-ability data
func (m *IOStreamsMock) FinishLogFile(ctx context.Context) {}

func (m *IOStreamsMock) FlushToLogFile(ctx context.Context, prefix, errStr string) error { return nil }

func (m *IOStreamsMock) FlushToLogstash(ctx context.Context) error { return nil }
