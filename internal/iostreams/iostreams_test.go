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
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IOSteams_NewIOStreams(t *testing.T) {
	var io *IOStreams
	fsMock := slackdeps.NewFsMock()
	osMock := slackdeps.NewOsMock()
	config := config.NewConfig(fsMock, osMock)
	config.DebugEnabled = true
	io = NewIOStreams(config, fsMock, osMock)
	require.True(t, io.config.DebugEnabled, "iostreams references config")
}

func Test_IOStreams_IsTTY(t *testing.T) {
	tests := map[string]struct {
		fileInfo os.FileInfo
		expected bool
	}{
		"interactive when stdout is a char device": {
			fileInfo: &slackdeps.FileInfoCharDevice{},
			expected: true,
		},
		"not interactive for different file modes": {
			fileInfo: &slackdeps.FileInfoNamedPipe{},
			expected: false,
		},
		"errors when checking interactivity fallback to false": {
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			config := config.NewConfig(fsMock, osMock)
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: tt.fileInfo})
			io := NewIOStreams(config, fsMock, osMock)

			isTTY := io.IsTTY()
			assert.Equal(t, isTTY, tt.expected)
		})
	}
}
