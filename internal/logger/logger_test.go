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

package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CmdClientMock struct {
	mock.Mock
}

func (m *CmdClientMock) OnSuccess(event *LogEvent) {
	m.Called()
}

func (m *CmdClientMock) OnEvent(event *LogEvent) {
	m.Called()
}

func TestLogger(t *testing.T) {
	cmdClientMock := new(CmdClientMock)
	cmdClientMock.On("OnEvent").Return()

	log := New(cmdClientMock.OnEvent)

	log.Log("info", "app_create")
	cmdClientMock.AssertCalled(t, "OnEvent")
}

func TestLoggerDebug(t *testing.T) {
	cmdClientMock := new(CmdClientMock)
	cmdClientMock.On("OnEvent").Return()

	log := New(cmdClientMock.OnEvent)
	log.Debug("debug_event")
	cmdClientMock.AssertCalled(t, "OnEvent")
}

func TestLoggerInfo(t *testing.T) {
	cmdClientMock := new(CmdClientMock)
	cmdClientMock.On("OnEvent").Return()

	log := New(cmdClientMock.OnEvent)
	log.Info("info_event")
	cmdClientMock.AssertCalled(t, "OnEvent")
}

func TestLoggerWarn(t *testing.T) {
	cmdClientMock := new(CmdClientMock)
	cmdClientMock.On("OnEvent").Return()

	log := New(cmdClientMock.OnEvent)
	log.Warn("warn_event")
	cmdClientMock.AssertCalled(t, "OnEvent")
}

func TestLoggerSuccessEvent(t *testing.T) {
	log := New(nil)
	log.Data = LogData{"key": "value"}
	event := log.SuccessEvent()
	assert.Equal(t, "info", event.Level)
	assert.Equal(t, "success", event.Name)
	assert.Equal(t, log.Data, event.Data)
}

func TestLoggerLogWithNilHandler(t *testing.T) {
	log := New(nil)
	// Should not panic when onEvent is nil
	log.Log("info", "no_handler")
}
