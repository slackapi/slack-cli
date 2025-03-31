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
	m.On("FlushToLogstash", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func (m *EventTrackerMock) FlushToLogstash(ctx context.Context, cfg config.Config, io iostreams.IOStreamer, exitCode iostreams.ExitCode) error {
	args := m.Called(ctx)
	return args.Error(0)
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
