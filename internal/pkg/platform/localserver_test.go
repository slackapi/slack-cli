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

package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var WebsocketDialerDial = &websocketDialerDial

func Test_LocalServer_Start(t *testing.T) {
	for name, tt := range map[string]struct {
		Setup      func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock)
		Test       func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock)
		wsHandler  func(w http.ResponseWriter, r *http.Request)
		fakeDialer func(conn *WebSocketConnMock) func(d *websocket.Dialer, urlStr string,
			requestHeader http.Header) (WebSocketConnection, *http.Response, error)
	}{
		"should return an error if there was a problem asking for a WebSocket connection URL": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				cm.APIInterface.On("ConnectionsOpen", mock.Anything, mock.Anything).Return(api.AppsConnectionsOpenResult{}, slackerror.New("no can do, pipes are clogged"))
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				require.ErrorContains(t, server.Start(ctx), "pipes are clogged")
			},
		},
		"should return an error if there was a problem connecting to a WebSocket connection URL": {
			wsHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				require.ErrorContains(t, server.Start(ctx), "bad handshake")
			},
		},
		"should return an error if there was a problem reading from WebSocket": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(0, []byte{}, slackerror.New("oh no"))
			},
			fakeDialer: func(conn *WebSocketConnMock) func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
				return func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
					return conn, nil, nil
				}
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				require.ErrorContains(t, server.Start(ctx), "oh no")
			},
		},
		"should re-establish connection if disconnect message received": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"disconnect\"}"), nil).Once()
				conn.On("ReadMessage").Return(websocket.CloseMessage, []byte{}, &websocket.CloseError{Code: websocket.CloseNormalClosure, Text: "byebye"}).Once()
			},
			fakeDialer: func(conn *WebSocketConnMock) func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
				return func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
					return conn, nil, nil
				}
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				require.ErrorContains(t, server.Start(ctx), slackerror.ErrLocalAppRunCleanExit)
				// Once to re-establish post-disconnect message and once when close message received
				conn.AssertNumberOfCalls(t, "Close", 2)
			},
		},
		"should re-establish connection if unexpected, non-JSON response received": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				cm.Config.DebugEnabled = true
				// Simulate TWO bad messages received, followed by a close message
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("CANCELLED: Failed to read message"), nil).Once()
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("CANCELLED: Failed to read message"), nil).Once()
				conn.On("ReadMessage").Return(websocket.CloseMessage, []byte{}, &websocket.CloseError{Code: websocket.CloseNormalClosure, Text: "byebye"}).Once()
			},
			fakeDialer: func(conn *WebSocketConnMock) func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
				return func(d *websocket.Dialer, urlStr string, requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
					return conn, nil, nil
				}
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				require.ErrorContains(t, server.Start(ctx), slackerror.ErrLocalAppRunCleanExit)
				// Expectation is each WS message we configured in Setup
				// would cause the TCP connection to be closed (3 times)
				// but the connection should be re-established only twice (once for each simulated error)
				conn.AssertNumberOfCalls(t, "Close", 3)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			wsFakeURL := "http://slack.com/websocketzzzz"
			if tt.wsHandler != nil {
				ts := httptest.NewServer(http.HandlerFunc(tt.wsHandler))
				defer ts.Close()
				wsFakeURL = "ws" + strings.TrimPrefix(ts.URL, "http")
			}
			// Create mocks
			ctx := slackcontext.MockContext(t.Context())
			conn := NewWebSocketConnMock()
			clientsMock := shared.NewClientsMock()
			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
			})
			if tt.Setup != nil {
				tt.Setup(t, clientsMock, clients, conn)
			}
			clientsMock.APIInterface.On("ConnectionsOpen", mock.Anything, mock.Anything).Return(api.AppsConnectionsOpenResult{URL: wsFakeURL}, nil)
			// Setup default mock actions
			conn.AddDefaultMocks()
			clientsMock.AddDefaultMocks()
			localContext := LocalHostedContext{
				BotAccessToken: "ABC1234",
				AppID:          "A12345",
				TeamID:         "justiceleague",
			}
			log := logger.Logger{
				Data: map[string]interface{}{},
			}
			server := LocalServer{
				clients,
				&log,
				"ABC123",
				localContext,
				clients.SDKConfig,
				conn,
			}
			if tt.fakeDialer != nil {
				orig := *WebsocketDialerDial
				websocketDialerDial = tt.fakeDialer(conn)
				defer func() {
					websocketDialerDial = orig
				}()
			}

			tt.Test(t, ctx, clientsMock, server, conn)
		})
	}
}

func Test_LocalServer_Listen(t *testing.T) {
	for name, tt := range map[string]struct {
		Setup func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock)
		Test  func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock)
	}{
		"should return and send special clean exit error if context is canceled": {
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				ctx2, cancel := context.WithCancel(ctx)
				cancel()
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx2, errChan, done)
				select {
				case err := <-errChan:
					assert.True(t, slackerror.Is(err, slackerror.ErrLocalAppRunCleanExit))
				case <-done:
					assert.Fail(t, "unexpected done channel signalled")
				}
			},
		},
		"should return and send an error if ReadMessage has an unexpected error": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(0, []byte{}, slackerror.New("oh no"))
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.ErrorContains(t, err, "oh no")
				case <-done:
					assert.Fail(t, "unexpected done channel signalled")
				}
			},
		},
		"should return and send special clean exit error if ReadMessage receives a normal closure message": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(websocket.CloseMessage, []byte{}, &websocket.CloseError{Code: websocket.CloseNormalClosure, Text: "byebye"})
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.True(t, slackerror.Is(err, slackerror.ErrLocalAppRunCleanExit))
				case <-done:
					assert.Fail(t, "unexpected done channel signalled")
				}
			},
		},
		"should signal done=false (signalling a 'plz re-restablish socket connection') if we receive a non-JSON message": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("cache_error"), nil)
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.Nil(t, err)
				case quit := <-done:
					assert.False(t, quit)
				}
			},
		},
		"should return and send on done if a disconnect message type is received": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"disconnect\"}"), nil)
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.Fail(t, "unexpected err channel signalled", err)
				case quit := <-done:
					assert.False(t, quit)
				}
			},
		},
		"should return and send an error if there was a problem retrieving the SDK start hook": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"event\",\"payload\":{}}"), nil)
				// Force the start hook to return an error
				// TODO: should probably create a hookscript mock instead of doing this.
				clients.SDKConfig.Hooks.Start = hooks.HookScript{Command: "", Name: "start"}
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.ErrorContains(t, err, "command for 'start' was not found")
				case <-done:
					assert.Fail(t, "unexpected done channel signalled")
				}
			},
		},
		"should send a websocket response message if socket event message received was passed to start hook and hook returns successfully": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				// TODO: should probably create a hookscript mock instead of doing this.
				clients.SDKConfig.Hooks.Start = hooks.HookScript{Command: "echo '{}'", Name: "start"}
				// Simulate receiving an event, then a disconnect message
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"event\",\"payload\":{}}"), nil).Once()
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"disconnect\"}"), nil).Once()
				cm.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return("{}", nil)
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.Fail(t, "unexpected err channel signalled", err)
				case <-done:
					conn.AssertNumberOfCalls(t, "WriteMessage", 1)
				}
			},
		},
		"should not send a websocket response message if socket event message received led to start hook error": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				// TODO: should probably create a hookscript mock instead of doing this.
				clients.SDKConfig.Hooks.Start = hooks.HookScript{Command: "echo '{}'", Name: "start"}
				// Simulate receiving an event, then a disconnect message
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"event\",\"payload\":{}}"), nil).Once()
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"disconnect\"}"), nil).Once()
				cm.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return("{}", slackerror.New("typescript error, probably"))
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.Fail(t, "unexpected err channel signalled", err)
				case <-done:
					conn.AssertNumberOfCalls(t, "WriteMessage", 0)
				}
			},
		},
		"should return and send an error if there was a problem sending a websocket message": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, clients *shared.ClientFactory, conn *WebSocketConnMock) {
				// TODO: should probably create a hookscript mock instead of doing this.
				clients.SDKConfig.Hooks.Start = hooks.HookScript{Command: "echo '{}'", Name: "start"}
				// Simulate receiving an event, then a disconnect message
				conn.On("ReadMessage").Return(websocket.TextMessage, []byte("{\"type\":\"event\",\"payload\":{}}"), nil).Once()
				cm.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return("{}", nil)
				conn.On("WriteMessage", mock.Anything, mock.Anything).Return(slackerror.New("socket pipe severed"))
			},
			Test: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, server LocalServer, conn *WebSocketConnMock) {
				errChan := make(chan error)
				done := make(chan bool)
				go server.Listen(ctx, errChan, done)
				select {
				case err := <-errChan:
					assert.ErrorContains(t, err, "socket pipe severed")
				case <-done:
					assert.Fail(t, "unexpected done channel signalled")
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			ctx := slackcontext.MockContext(t.Context())
			conn := NewWebSocketConnMock()
			clientsMock := shared.NewClientsMock()
			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
			})
			if tt.Setup != nil {
				tt.Setup(t, clientsMock, clients, conn)
			}
			// Setup default mock actions
			conn.AddDefaultMocks()
			clientsMock.AddDefaultMocks()
			localContext := LocalHostedContext{
				BotAccessToken: "ABC1234",
				AppID:          "A12345",
				TeamID:         "justiceleague",
			}
			log := logger.Logger{
				Data: map[string]interface{}{},
			}
			server := LocalServer{
				clients,
				&log,
				"ABC123",
				localContext,
				clients.SDKConfig,
				conn,
			}
			tt.Test(t, ctx, clientsMock, server, conn)
		})
	}
}

func Test_sendWebSocketMessage(t *testing.T) {
	t.Run("should error when linkResponse param is nil", func(t *testing.T) {
		conn := NewWebSocketConnMock()
		conn.AddDefaultMocks()

		err := sendWebSocketMessage(conn, nil)
		require.Error(t, err)
		conn.AssertNotCalled(t, "WriteMessage", mock.Anything, mock.Anything)
	})

	t.Run("should write a message to the connection even when payload param is empty JSON", func(t *testing.T) {
		conn := NewWebSocketConnMock()
		conn.AddDefaultMocks()

		linkResponse := LinkResponse{
			EnvelopeID: validEnvelopeIDMock,
			Payload:    json.RawMessage("{}"),
		}

		expectedMessageData, err := json.Marshal(linkResponse)
		require.NoError(t, err)

		err = sendWebSocketMessage(conn, &linkResponse)
		require.NoError(t, err)
		conn.AssertCalled(t, "WriteMessage", websocket.TextMessage, expectedMessageData)
	})

	// When payload param is populated JSON
	t.Run("When payload param is populated JSON", func(t *testing.T) {
		conn := NewWebSocketConnMock()
		conn.AddDefaultMocks()

		linkResponse := LinkResponse{
			EnvelopeID: validEnvelopeIDMock,
			Payload:    json.RawMessage("{ \"hello\": \"world\" }"),
		}

		expectedMessageData, err := json.Marshal(linkResponse)
		require.NoError(t, err)

		err = sendWebSocketMessage(conn, &linkResponse)
		require.NoError(t, err)
		conn.AssertCalled(t, "WriteMessage", websocket.TextMessage, expectedMessageData)
	})

	t.Run("should error when payload param is invalid JSON", func(t *testing.T) {
		conn := NewWebSocketConnMock()
		conn.AddDefaultMocks()

		linkResponse := LinkResponse{
			EnvelopeID: validEnvelopeIDMock,
			Payload:    json.RawMessage("hello world"),
		}

		err := sendWebSocketMessage(conn, &linkResponse)
		require.Error(t, err)
		conn.AssertNotCalled(t, "WriteMessage", mock.Anything, mock.Anything)
	})

	t.Run("should error when writing to websocket fails", func(t *testing.T) {
		expectedError := fmt.Errorf("Error writing to websocket")
		conn := NewWebSocketConnMock()
		conn.On("WriteMessage", mock.Anything, mock.Anything).Return(expectedError)
		conn.AddDefaultMocks()

		linkResponse := LinkResponse{
			EnvelopeID: validEnvelopeIDMock,
			Payload:    json.RawMessage("{}"),
		}

		err := sendWebSocketMessage(conn, &linkResponse)
		require.ErrorIs(t, err, expectedError)
		conn.AssertCalled(t, "WriteMessage", mock.Anything, mock.Anything)
	})
}
