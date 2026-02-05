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

package iostreams

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/style"
)

// PrintDebug mocks printing a message to stdout only when debug is enabled
func (m *IOStreamsMock) PrintDebug(ctx context.Context, format string, a ...interface{}) {
	m.Called(ctx, format, a)
	errStr := fmt.Sprintf(format, a...)
	if m.config.DebugEnabled {
		m.Stdout.Println(style.Secondary((errStr)))
	}
}

// PrintError mocks logging and printing an error message to stderr
func (m *IOStreamsMock) PrintError(ctx context.Context, format string, a ...interface{}) {
	err := fmt.Sprintf(format, a...)
	m.Stderr.Println(err)
}

// PrintWarning falsely mocks logs and prints a warning message to stdout
func (m *IOStreamsMock) PrintWarning(ctx context.Context, format string, a ...interface{}) {
	m.Called(ctx, format, a)
	err := fmt.Sprintf(format, a...)
	m.Stdout.Println(err)
}

// PrintInfo print a formatted message to stdout, sometimes tracing context
func (m *IOStreamsMock) PrintInfo(ctx context.Context, shouldTrace bool, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	m.Stdout.Println(msg)
}

// PrintTrace mocks how traces are generated and printed, matching the actual implementation
func (m *IOStreamsMock) PrintTrace(ctx context.Context, traceID string, traceValues ...string) {
	m.Called(ctx, traceID, traceValues)
	if !m.config.SlackTestTraceFlag {
		return
	}
	m.Stdout.Println(style.Tracef(traceID, traceValues...))
}
