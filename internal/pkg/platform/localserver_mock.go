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
	"time"

	"github.com/stretchr/testify/mock"
)

// validEnvelopeIDMock is an example of a valid LinkResponse.EnvelopeID for testing.
const validEnvelopeIDMock = "dbdd0ef3-1543-4f94-bfb4-133d0e6c1545"

// NewWebSocketConnMock creates a mock of WebSocketConnMock.
func NewWebSocketConnMock() *WebSocketConnMock {
	return &WebSocketConnMock{}
}

// WebSocketConnMock mocks the github.com/gorilla/websocket websocket.Conn struct.
type WebSocketConnMock struct {
	mock.Mock
}

// AddDefaultMocks installs the default mock actions to fallback on.
func (m *WebSocketConnMock) AddDefaultMocks() {
	m.On("WriteMessage", mock.Anything, mock.Anything).Return(nil)
	m.On("WriteControl", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.On("ReadMessage").Return(1 /* WebSocket TextMessage message type */, []byte("{}"), nil)
	m.On("Close").Return(nil)
}

// WriteMessage mock
func (m *WebSocketConnMock) WriteMessage(messageType int, data []byte) error {
	args := m.Called(messageType, data)
	return args.Error(0)
}

// ReadMessage mock
func (m *WebSocketConnMock) ReadMessage() (int, []byte, error) {
	args := m.Called()
	return args.Int(0), args.Get(1).([]byte), args.Error(2)
}

// WriteControl mock
func (m *WebSocketConnMock) WriteControl(messageType int, data []byte, deadline time.Time) error {
	args := m.Called(messageType, data, deadline)
	return args.Error(0)
}

// Close mock
func (m *WebSocketConnMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
