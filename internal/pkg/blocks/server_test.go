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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_buildBlockKitBuilderURL(t *testing.T) {
	tests := map[string]struct {
		apiHost    string
		teamID     string
		port       int
		blocksJSON string
		expected   []string
	}{
		"constructs correct URL for dev instance": {
			apiHost:    "https://dev1388.slack.com",
			teamID:     "T0123456789",
			port:       12345,
			blocksJSON: `{"blocks":[]}`,
			expected: []string{
				"app.dev1388.slack.com/block-kit-builder/T0123456789/builder",
				"ws_port=12345",
				"%7B%22blocks%22:%5B%5D%7D",
			},
		},
		"constructs correct URL for production": {
			apiHost:    "https://slack.com",
			teamID:     "T0123456789",
			port:       8080,
			blocksJSON: `{"blocks":[{"type":"section"}]}`,
			expected: []string{
				"app.slack.com/block-kit-builder/T0123456789/builder",
				"ws_port=8080",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := buildBlockKitBuilderURL(tc.apiHost, tc.teamID, tc.port, tc.blocksJSON)
			for _, exp := range tc.expected {
				assert.Contains(t, result, exp)
			}
		})
	}
}

func Test_Preview_ConnectionTimeout(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Return()

	originalTimeout := connectionTimeout
	connectionTimeout = 100 * time.Millisecond
	defer func() { connectionTimeout = originalTimeout }()

	ctx := t.Context()
	_, err := Preview(ctx, clients, "T0123456789", `{"blocks":[]}`, "/tmp/test.png")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Timed out")
}

func Test_Preview_ContextCancelled(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Return()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := Preview(ctx, clients, "T0123456789", `{"blocks":[]}`, "/tmp/test.png")
	assert.Error(t, err)
}

func Test_Preview_ListenError(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	originalListen := netListen
	netListen = func(network, address string) (net.Listener, error) {
		return nil, fmt.Errorf("bind failed")
	}
	defer func() { netListen = originalListen }()

	ctx := t.Context()
	_, err := Preview(ctx, clients, "T0123456789", `{"blocks":[]}`, "/tmp/test.png")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bind failed")
}

func Test_Preview_Success(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	blocksJSON := `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"Hello"}}]}`
	teamID := "T0123456789"
	fakePNG := []byte{0x89, 0x50, 0x4E, 0x47}

	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Run(func(args mock.Arguments) {
		openedURL := args.Get(0).(string)
		assert.Contains(t, openedURL, "app.slack.com/block-kit-builder/T0123456789/builder")
		assert.Contains(t, openedURL, "ws_port=")

		go func() {
			time.Sleep(50 * time.Millisecond)
			portStr := extractWSPort(openedURL)
			wsURL := fmt.Sprintf("ws://127.0.0.1:%s/", portStr)
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send CONNECTED
			cp, _ := json.Marshal(connectedPayload{Version: "1.0.0"})
			connMsg, _ := json.Marshal(wsMessage{Type: "CONNECTED", Payload: cp})
			_ = conn.WriteMessage(websocket.TextMessage, connMsg)

			// Read REQUEST_SCREENSHOT
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req wsMessage
			_ = json.Unmarshal(data, &req)
			assert.Equal(t, "REQUEST_SCREENSHOT", req.Type)

			// Send SCREENSHOT response
			payload, _ := json.Marshal(screenshotPayload{
				Image:  base64.StdEncoding.EncodeToString(fakePNG),
				Width:  620,
				Height: 400,
			})
			resp, _ := json.Marshal(wsMessage{
				Type:    "SCREENSHOT",
				Payload: payload,
			})
			_ = conn.WriteMessage(websocket.TextMessage, resp)
		}()
	}).Return()

	ctx := t.Context()
	outputPath := filepath.Join(slackdeps.MockHomeDirectory, ".slack", "previews", "blocks-preview.png")
	filePath, err := Preview(ctx, clients, teamID, blocksJSON, outputPath)

	assert.NoError(t, err)
	assert.Equal(t, outputPath, filePath)

	data, err := afero.ReadFile(clients.Fs, filePath)
	assert.NoError(t, err)
	assert.Equal(t, fakePNG, data)
}

func Test_Preview_ErrorResponse(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Run(func(args mock.Arguments) {
		openedURL := args.Get(0).(string)

		go func() {
			time.Sleep(50 * time.Millisecond)
			portStr := extractWSPort(openedURL)
			wsURL := fmt.Sprintf("ws://127.0.0.1:%s/", portStr)
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send CONNECTED
			cp, _ := json.Marshal(connectedPayload{Version: "1.0.0"})
			connMsg, _ := json.Marshal(wsMessage{Type: "CONNECTED", Payload: cp})
			_ = conn.WriteMessage(websocket.TextMessage, connMsg)

			// Read REQUEST_SCREENSHOT
			_, _, _ = conn.ReadMessage()

			// Send ERROR response
			payload, _ := json.Marshal(errorPayload{
				Message: "Preview card element not found",
				Code:    "SCREENSHOT_FAILED",
			})
			resp, _ := json.Marshal(wsMessage{
				Type:    "ERROR",
				Payload: payload,
			})
			_ = conn.WriteMessage(websocket.TextMessage, resp)
		}()
	}).Return()

	ctx := t.Context()
	_, err := Preview(ctx, clients, "T0123456789", `{"blocks":[]}`, "/tmp/test.png")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Preview card element not found")
}

func Test_Preview_ResponseTimeout(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.Browser.ExpectedCalls = nil
	clientsMock.Browser.On("OpenURL", mock.Anything).Run(func(args mock.Arguments) {
		openedURL := args.Get(0).(string)

		go func() {
			time.Sleep(50 * time.Millisecond)
			portStr := extractWSPort(openedURL)
			wsURL := fmt.Sprintf("ws://127.0.0.1:%s/", portStr)
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send CONNECTED
			cp, _ := json.Marshal(connectedPayload{Version: "1.0.0"})
			connMsg, _ := json.Marshal(wsMessage{Type: "CONNECTED", Payload: cp})
			_ = conn.WriteMessage(websocket.TextMessage, connMsg)

			// Read REQUEST_SCREENSHOT but never respond
			_, _, _ = conn.ReadMessage()
			time.Sleep(500 * time.Millisecond)
		}()
	}).Return()

	originalTimeout := responseTimeout
	responseTimeout = 100 * time.Millisecond
	defer func() { responseTimeout = originalTimeout }()

	ctx := t.Context()
	_, err := Preview(ctx, clients, "T0123456789", `{"blocks":[]}`, "/tmp/test.png")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Timed out waiting for screenshot")
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
