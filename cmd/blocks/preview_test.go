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

package blocks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewCommand(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	assert.True(t, cmd.Hidden)
	assert.Equal(t, "blocks", cmd.Name())
}

func Test_NewPreviewCommand(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewPreviewCommand(clients)
	assert.Equal(t, "preview", cmd.Name())
}

func Test_PreviewCommand_MissingInput(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789", "--output", "/tmp/out.png"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No blocks JSON provided")
}

func Test_PreviewCommand_InvalidJSON(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.IO.Stdin = strings.NewReader("{not valid json")

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789", "--output", "/tmp/out.png"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "looking for beginning of object key string")
}

func Test_PreviewCommand_Stdin(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	blocksJSON := `{"blocks":[]}`
	fakePNG := []byte{0x89, 0x50, 0x4E, 0x47}
	customOutput := "/tmp/stdin-preview.png"

	clientsMock.IO.Stdin = strings.NewReader(blocksJSON)
	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Run(func(args mock.Arguments) {
		openedURL := args.Get(0).(string)
		go simulateBlockKitBuilder(openedURL, blocksJSON, fakePNG)
	}).Return()

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789", "--output", customOutput})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.NoError(t, err)

	output := clientsMock.GetCombinedOutput()
	assert.Contains(t, output, customOutput)
}

func Test_PreviewCommand_MissingTeamFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.IO.Stdin = strings.NewReader(`{"blocks":[]}`)

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Team ID is required")
}

func Test_PreviewCommand_OutputFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	blocksJSON := `{"blocks":[]}`
	fakePNG := []byte{0x89, 0x50, 0x4E, 0x47}
	customOutput := "/tmp/my-preview.png"

	clientsMock.IO.Stdin = strings.NewReader(blocksJSON)
	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Run(func(args mock.Arguments) {
		openedURL := args.Get(0).(string)
		go simulateBlockKitBuilder(openedURL, blocksJSON, fakePNG)
	}).Return()

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789", "--output", customOutput})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.NoError(t, err)

	output := clientsMock.GetCombinedOutput()
	assert.Contains(t, output, customOutput)
}

func Test_PreviewCommand_MissingOutputFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.IO.Stdin = strings.NewReader(`{"blocks":[]}`)

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Output file path is required")
}

func Test_compactJSON(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
		wantErr  bool
	}{
		"already compact": {
			input:    `{"blocks":[]}`,
			expected: `{"blocks":[]}`,
		},
		"removes whitespace": {
			input:    "{\n  \"blocks\": [\n    {\n      \"type\": \"section\"\n    }\n  ]\n}",
			expected: `{"blocks":[{"type":"section"}]}`,
		},
		"invalid JSON returns error": {
			input:   "{not valid",
			wantErr: true,
		},
		"empty string returns error": {
			input:   "",
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := compactJSON(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_validateBlocksPayload(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid blocks payload": {
			input: `{"blocks":[]}`,
		},
		"valid with content": {
			input: `{"blocks":[{"type":"section"}]}`,
		},
		"extra fields alongside blocks is valid": {
			input: `{"blocks":[],"metadata":"x"}`,
		},
		"missing blocks key returns error": {
			input:   `{"type":"section"}`,
			wantErr: true,
		},
		"blocks field is not an array returns error": {
			input:   `{"blocks":"hello"}`,
			wantErr: true,
		},
		"blocks field is an object returns error": {
			input:   `{"blocks":{}}`,
			wantErr: true,
		},
		"blocks field is null returns error": {
			input:   `{"blocks":null}`,
			wantErr: true,
		},
		"top-level array returns error": {
			input:   `[{"type":"section"}]`,
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateBlocksPayload(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func simulateBlockKitBuilder(openedURL string, _ string, imageBytes []byte) {
	time.Sleep(50 * time.Millisecond)
	portStr := extractWSPort(openedURL)
	wsURL := fmt.Sprintf("ws://127.0.0.1:%s/", portStr)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	type msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	// Send CONNECTED
	cp, _ := json.Marshal(map[string]string{"version": "1.0.0"})
	connMsg, _ := json.Marshal(msg{Type: "CONNECTED", Payload: cp})
	_ = conn.WriteMessage(websocket.TextMessage, connMsg)

	// Read REQUEST_SCREENSHOT
	_, _, _ = conn.ReadMessage()

	// Send SCREENSHOT response
	type screenshotPayload struct {
		Image  string `json:"image"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	p, _ := json.Marshal(screenshotPayload{
		Image:  base64.StdEncoding.EncodeToString(imageBytes),
		Width:  620,
		Height: 400,
	})
	resp, _ := json.Marshal(msg{Type: "SCREENSHOT", Payload: p})
	_ = conn.WriteMessage(websocket.TextMessage, resp)
}

func extractWSPort(builderURL string) string {
	idx := strings.Index(builderURL, "ws_port=")
	if idx == -1 {
		return ""
	}
	rest := builderURL[idx+len("ws_port="):]
	end := strings.IndexAny(rest, "&#")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
