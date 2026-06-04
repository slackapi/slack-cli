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
	"time"

	"github.com/stretchr/testify/mock"
)

type wsConnMock struct {
	mock.Mock
}

func newWSConnMock() *wsConnMock {
	return &wsConnMock{}
}

func (m *wsConnMock) AddDefaultMocks() {
	m.On("readMessage", mock.Anything).Return(wsMessage{}, nil)
	m.On("writeMessage", mock.Anything).Return(nil)
	m.On("Close").Return(nil)
}

func (m *wsConnMock) readMessage(timeout time.Duration) (wsMessage, error) {
	args := m.Called(timeout)
	return args.Get(0).(wsMessage), args.Error(1)
}

func (m *wsConnMock) writeMessage(msg wsMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *wsConnMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
