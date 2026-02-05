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

package tracking

import (
	"context"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/stretchr/testify/mock"
)

type EventTrackerMock struct {
	mock.Mock
}

func (m *EventTrackerMock) AddDefaultMocks() {
	m.On("SetAppEnterpriseID", mock.Anything)
	m.On("SetAppTeamID", mock.Anything)
	m.On("SetAppTemplate", mock.Anything)
	m.On("SetAppUserID", mock.Anything)
	m.On("SetAuthEnterpriseID", mock.Anything)
	m.On("SetAuthTeamID", mock.Anything)
	m.On("SetAuthUserID", mock.Anything)
	m.On("SetErrorCode", mock.Anything)
	m.On("SetErrorMessage", mock.Anything)
}

func (m *EventTrackerMock) FlushToLogstash(ctx context.Context, cfg *config.Config, io iostreams.IOStreamer, exitCode iostreams.ExitCode) error {
	args := m.Called(ctx, cfg, io, exitCode)
	return args.Error(0)
}

func (m *EventTrackerMock) getSessionData() EventData {
	args := m.Called()
	return args.Get(0).(EventData)
}

func (m *EventTrackerMock) cleanSessionData(data EventData) EventData {
	args := m.Called(data)
	return args.Get(0).(EventData)
}

func (m *EventTrackerMock) SetErrorCode(code string) {
	m.Called(code)
}

func (m *EventTrackerMock) SetErrorMessage(err string) {
	m.Called(err)
}

func (m *EventTrackerMock) SetAuthEnterpriseID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAuthTeamID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAuthUserID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAppEnterpriseID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAppTeamID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAppUserID(id string) {
	m.Called(id)
}

func (m *EventTrackerMock) SetAppTemplate(template string) {
	m.Called(template)
}
