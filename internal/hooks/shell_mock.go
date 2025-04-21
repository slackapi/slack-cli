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
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// Mock that implements ExecInterface for testing
type MockExec struct {
	mockCommand *MockCommand
}

func (e *MockExec) Command(env []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, name string, arg ...string) ShellCommand {
	e.mockCommand.Env = env
	e.mockCommand.stdout = stdout
	e.mockCommand.stderr = stderr
	return e.mockCommand
}

type MockCommand struct {
	MockStdout []byte
	MockStderr []byte
	Err        error
	Env        []string
	Args       map[string]string
	StdoutIO   io.ReadCloser
	StderrIO   io.ReadCloser

	stdout io.Writer
	stderr io.Writer
}

func (c *MockCommand) Run() error {
	if len(c.MockStdout) > 0 {
		_, _ = c.stdout.Write(c.MockStdout)
	} else if c.StdoutIO != nil {
		_, _ = io.Copy(c.stdout, c.StdoutIO)
	}
	if len(c.MockStderr) > 0 {
		_, _ = c.stderr.Write(c.MockStderr)
	} else if c.StderrIO != nil {
		_, _ = io.Copy(c.stderr, c.StderrIO)
	}
	return c.Err
}

func (c *MockCommand) Start() error {
	return c.Err
}

func (c *MockCommand) Wait() error {
	return c.Err
}

func (c *MockCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.StdoutIO, c.Err
}

func (c *MockCommand) StderrPipe() (io.ReadCloser, error) {
	return c.StderrIO, c.Err
}

func (c *MockCommand) CombinedOutput() ([]byte, error) {
	return c.MockStdout, c.Err
}

type MockHookExecutor struct {
	mock.Mock
}

func (m *MockHookExecutor) Execute(ctx context.Context, opts HookExecOpts) (string, error) {
	args := m.Called(ctx, opts)
	return args.String(0), args.Error(1)
}
