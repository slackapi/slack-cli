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
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BoundariedWriter(t *testing.T) {
	tests := map[string]struct {
		bw           BoundariedWriter
		writes       []string
		expectedBuff string
	}{
		"outputs between bounds are buffered while all are streamed": {
			bw: BoundariedWriter{
				Buff:   &bytes.Buffer{},
				Bounds: "xoxo",
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"hello world!" + EOL,
				`xoxo{"hooks":"example"}xoxo` + EOL,
				"farewell o7" + EOL,
			},
			expectedBuff: `{"hooks":"example"}`,
		},
		"multiline output between bounds are parsed": {
			bw: BoundariedWriter{
				Buff:   &bytes.Buffer{},
				Bounds: "xoxo",
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"hello world!" + EOL,
				`xoxo{"hooks":"ex` + EOL,
				"amp" + EOL,
				`le"}xoxo` + EOL,
				"farewell o7" + EOL,
			},
			expectedBuff: `{"hooks":"example"}`,
		},
		"inlined bounds are parsed": {
			bw: BoundariedWriter{
				Buff:   &bytes.Buffer{},
				Bounds: "314",
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				`greetings314{"hooks":"inline"}314continues`,
			},
			expectedBuff: `{"hooks":"inline"}`,
		},
		"missing buffers still output write statements to the stream": {
			bw: BoundariedWriter{
				Bounds: "...",
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"...-...",
			},
		},
		"missing streams still output write statements to the buffer": {
			bw: BoundariedWriter{
				Buff:   &bytes.Buffer{},
				Bounds: "**",
			},
			writes: []string{
				"recall what's **bold**",
			},
			expectedBuff: "bold",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, line := range tc.writes {
				n, err := tc.bw.Write([]byte(line))
				require.NoError(t, err)
				require.Equal(t, len(line), n)
			}
			if tc.bw.Buff != nil {
				assert.Equal(t, tc.expectedBuff, tc.bw.Buff.(*bytes.Buffer).String())
			}
			if tc.bw.Stream != nil {
				assert.Equal(t, strings.Join(tc.writes, ""), tc.bw.Stream.(*bytes.Buffer).String())
			}
		})
	}
}

func Test_BufferedWriter(t *testing.T) {
	tests := map[string]struct {
		bw     BufferedWriter
		writes []string
	}{
		"matching outputs are written to different writers": {
			bw: BufferedWriter{
				Buff:   &bytes.Buffer{},
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"beepboop",
				"robot sounds",
			},
		},
		"missing buffers still output write statements to the stream": {
			bw: BufferedWriter{
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"wavecrash",
				"ocean sounds",
			},
		},
		"missing streams still output write statements to the buffer": {
			bw: BufferedWriter{
				Buff: &bytes.Buffer{},
			},
			writes: []string{
				"waxon",
				"and repeat",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, line := range tc.writes {
				n, err := tc.bw.Write([]byte(line))
				require.NoError(t, err)
				require.Equal(t, len(line), n)
			}
			if tc.bw.Buff != nil {
				assert.Equal(t, strings.Join(tc.writes, ""), tc.bw.Buff.(*bytes.Buffer).String())
			}
			if tc.bw.Stream != nil {
				assert.Equal(t, strings.Join(tc.writes, ""), tc.bw.Stream.(*bytes.Buffer).String())
			}
		})
	}
}

func Test_FilteredWriter(t *testing.T) {
	tests := map[string]struct {
		fw             FilteredWriter
		writes         []string
		expectedStream string
	}{
		"unfiltered outputs are written to the stream": {
			fw: FilteredWriter{
				Bounds: "xoxo",
				Stream: &bytes.Buffer{},
			},
			writes: []string{
				"beepboop\n",
				"xoxo{response}xoxo\n",
				"robot sounds\n",
			},
			expectedStream: "beepboop\nrobot sounds\n",
		},
		"missing streams attempt to write without errors": {
			fw: FilteredWriter{},
			writes: []string{
				"wavecrash",
				"ocean sounds",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, line := range tc.writes {
				n, err := tc.fw.Write([]byte(line))
				require.NoError(t, err)
				require.Equal(t, len(line), n)
			}
			if tc.fw.Stream != nil {
				assert.Equal(t, tc.expectedStream, tc.fw.Stream.(*bytes.Buffer).String())
			}
		})
	}
}

func Test_WriteIndent(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"outputs nothing if provided no input": {
			input:    "",
			expected: "",
		},
		"indents multiple lines with formatting": {
			input:    "xoxo\nxoxo\n",
			expected: "   xoxo\n   xoxo\n",
		},
		"prints a single line with indentation": {
			input:    "nice\n",
			expected: "   nice\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			config := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(config, fsMock, osMock)
			buff := &bytes.Buffer{}
			w := io.WriteIndent(buff)
			n, err := w.Write([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, len(tc.input), n)
			assert.Equal(t, tc.expected, buff.String())
		})
	}
}

func Test_WriteSecondary(t *testing.T) {
	tests := map[string]struct {
		input     string
		expected  string
		formatted bool
	}{
		"outputs nothing if provided no input": {
			input:    "",
			expected: "",
		},
		"formats multiple lines with formatting": {
			input:     "xoxo\nxoxo\n",
			expected:  "\x1b[38;5;244mxoxo\x1b[0m\n\x1b[38;5;244mxoxo\x1b[0m\n",
			formatted: true,
		},
		"prints a line without formatting": {
			input:     "nice\n",
			expected:  "nice\n",
			formatted: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			config := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(config, fsMock, osMock)
			buff := &bytes.Buffer{}
			style.ToggleStyles(tc.formatted)
			defer func() {
				style.ToggleStyles(false)
			}()
			w := io.WriteSecondary(buff)
			n, err := w.Write([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, len(tc.input), n)
			assert.Equal(t, tc.expected, buff.String())
		})
	}
}
