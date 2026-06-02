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

func Test_PreviewCommand_MissingArg(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func Test_PreviewCommand_MissingAppFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", `{"blocks":[]}`})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Team ID is required")
}

func Test_PreviewCommand_Success(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	blocksJSON := `{"blocks":[]}`
	fakePNG := []byte{0x89, 0x50, 0x4E, 0x47}

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
			type msg struct {
				Type    string          `json:"type"`
				Payload json.RawMessage `json:"payload"`
			}
			cp, _ := json.Marshal(map[string]string{"version": "1.0.0"})
			connMsg, _ := json.Marshal(msg{Type: "CONNECTED", Payload: cp})
			_ = conn.WriteMessage(websocket.TextMessage, connMsg)

			// Read SET_BLOCKS
			_, _, _ = conn.ReadMessage()

			// Send BLOCKS_UPDATED
			bup, _ := json.Marshal(map[string]interface{}{"json": blocksJSON, "success": true})
			updatedMsg, _ := json.Marshal(msg{Type: "BLOCKS_UPDATED", Payload: bup})
			_ = conn.WriteMessage(websocket.TextMessage, updatedMsg)

			// Read REQUEST_SCREENSHOT
			_, _, _ = conn.ReadMessage()

			// Send SCREENSHOT response
			type screenshotPayload struct {
				Image  string `json:"image"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			}
			p, _ := json.Marshal(screenshotPayload{
				Image:  base64.StdEncoding.EncodeToString(fakePNG),
				Width:  620,
				Height: 400,
			})
			resp, _ := json.Marshal(msg{Type: "SCREENSHOT", Payload: p})
			_ = conn.WriteMessage(websocket.TextMessage, resp)
		}()
	}).Return()

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"preview", "--team", "T0123456789", blocksJSON})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.NoError(t, err)

	output := clientsMock.GetCombinedOutput()
	assert.Contains(t, output, "blocks-preview-")
	assert.Contains(t, output, ".png")
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
