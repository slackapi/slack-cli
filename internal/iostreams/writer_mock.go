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
)

// WriteOut returns the mocked writer associated with stdout
func (m *IOStreamsMock) WriteOut() io.Writer {
	return m.Stdout.Writer()
}

// WriteErr returns the mocked writer associated with stderr
func (m *IOStreamsMock) WriteErr() io.Writer {
	return m.Stderr.Writer()
}

// WriteDebug implements the actual WriteDebug method with a mock call
func (m *IOStreamsMock) WriteDebug(ctx context.Context) WriteDebugger {
	m.Called(ctx)
	return WriteDebugger{ctx: ctx, io: m}
}

// WriteIndent stubs the implementation and uses a default mock
func (m *IOStreamsMock) WriteIndent(w io.Writer) WriteIndenter {
	args := m.Called(w)
	return args.Get(0).(WriteIndenter)
}

// WriteSecondary stubs the implementation and uses a default mock
func (m *IOStreamsMock) WriteSecondary(w io.Writer) WriteSecondarier {
	args := m.Called(w)
	return args.Get(0).(WriteSecondarier)
}
