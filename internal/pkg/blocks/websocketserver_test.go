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
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_handshake(t *testing.T) {
	tests := map[string]struct {
		readMsg wsMessage
		readErr error
		wantErr string
	}{
		"succeeds with CONNECTED message": {
			readMsg: wsMessage{
				Type:    "CONNECTED",
				Payload: json.RawMessage(`{"version":"1.0.0"}`),
			},
		},
		"fails with unexpected message type": {
			readMsg: wsMessage{Type: "UNKNOWN"},
			wantErr: "Unexpected message type: UNKNOWN",
		},
		"fails when read returns error": {
			readErr: fmt.Errorf("connection reset"),
			wantErr: "connection reset",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wsMock := newWSConnMock()
			wsMock.On("readMessage", mock.Anything).Return(tc.readMsg, tc.readErr)

			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			ctx := t.Context()

			err := handshake(ctx, clients.IO, wsMock)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
			wsMock.AssertExpectations(t)
		})
	}
}

func Test_requestScreenshot(t *testing.T) {
	screenshotPayloadJSON, _ := json.Marshal(screenshotPayload{
		Image:  "aW1hZ2VkYXRh",
		Width:  620,
		Height: 400,
	})
	errorPayloadJSON, _ := json.Marshal(errorPayload{
		Message: "Preview card element not found",
		Code:    "SCREENSHOT_FAILED",
	})

	tests := map[string]struct {
		writeErr    error
		readMsg     wsMessage
		readErr     error
		wantErr     string
		wantPayload screenshotPayload
	}{
		"succeeds with SCREENSHOT response": {
			readMsg: wsMessage{
				Type:    "SCREENSHOT",
				Payload: json.RawMessage(screenshotPayloadJSON),
			},
			wantPayload: screenshotPayload{
				Image:  "aW1hZ2VkYXRh",
				Width:  620,
				Height: 400,
			},
		},
		"fails when write returns error": {
			writeErr: fmt.Errorf("broken pipe"),
			wantErr:  "broken pipe",
		},
		"fails with ERROR response": {
			readMsg: wsMessage{
				Type:    "ERROR",
				Payload: json.RawMessage(errorPayloadJSON),
			},
			wantErr: "Preview card element not found",
		},
		"fails with unexpected response type": {
			readMsg: wsMessage{Type: "SOMETHING_ELSE"},
			wantErr: "Unexpected response type: SOMETHING_ELSE",
		},
		"fails when read returns error": {
			readErr: fmt.Errorf("i/o timeout"),
			wantErr: "i/o timeout",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wsMock := newWSConnMock()
			wsMock.On("writeMessage", mock.Anything).Return(tc.writeErr)
			if tc.writeErr == nil {
				wsMock.On("readMessage", mock.Anything).Return(tc.readMsg, tc.readErr)
			}

			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			ctx := t.Context()

			result, err := requestScreenshot(ctx, clients.IO, wsMock)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantPayload, result)
			}
			wsMock.AssertExpectations(t)
		})
	}
}

func Test_newWebSocketServer(t *testing.T) {
	tests := map[string]struct {
		listenFunc func(string, string) (net.Listener, error)
		wantErr    string
	}{
		"creates server successfully": {},
		"fails when listen returns error": {
			listenFunc: func(string, string) (net.Listener, error) {
				return nil, fmt.Errorf("address already in use")
			},
			wantErr: "address already in use",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.listenFunc != nil {
				originalListen := netListen
				netListen = tc.listenFunc
				defer func() { netListen = originalListen }()
			}

			server, err := newWebSocketServer()

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Greater(t, server.Port, 0)
				assert.NotNil(t, server.ConnChan)
				assert.NotNil(t, server.ErrChan)
				server.Shutdown()
			}
		})
	}
}

func Test_webSocketServer_Shutdown(t *testing.T) {
	server, err := newWebSocketServer()
	assert.NoError(t, err)

	server.Shutdown()

	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("shutdown did not complete in time")
	default:
	}
}
