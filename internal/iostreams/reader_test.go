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
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
)

func Test_ReadIn(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"returns the configured stdin reader": {
			input:    "test input",
			expected: "test input",
		},
		"returns an empty reader": {
			input:    "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)
			io.Stdin = strings.NewReader(tc.input)

			reader := io.ReadIn()

			buf := make([]byte, len(tc.input)+1)
			n, _ := reader.Read(buf)
			assert.Equal(t, tc.expected, string(buf[:n]))
		})
	}
}
