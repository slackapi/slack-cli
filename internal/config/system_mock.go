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

package config

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type SystemConfigMock struct {
	mock.Mock
}

func (m *SystemConfigMock) SetCustomConfigDirPath(customConfigDirPath string) {
	m.Called(customConfigDirPath)
}

func (m *SystemConfigMock) UserConfig(ctx context.Context) (*SystemConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemConfig), args.Error(1)
}

func (m *SystemConfigMock) SlackConfigDir(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) LogsDir(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) GetLastUpdateCheckedAt(ctx context.Context) (time.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *SystemConfigMock) SetLastUpdateCheckedAt(ctx context.Context, lastUpdateCheckedAt time.Time) (path string, err error) {
	args := m.Called(ctx, lastUpdateCheckedAt)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) InitSystemID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) GetSystemID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) SetSystemID(ctx context.Context, systemID string) (string, error) {
	args := m.Called(ctx, systemID)
	return args.String(0), args.Error(1)
}

func (m *SystemConfigMock) GetSurveyConfig(ctx context.Context, id string) (SurveyConfig, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(SurveyConfig), args.Error(1)
}

func (m *SystemConfigMock) SetSurveyConfig(ctx context.Context, id string, surveyConfig SurveyConfig) error {
	args := m.Called(ctx, id, surveyConfig)
	return args.Error(0)
}

func (m *SystemConfigMock) GetTrustUnknownSources(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *SystemConfigMock) SetTrustUnknownSources(ctx context.Context, value bool) error {
	args := m.Called(ctx, value)
	return args.Error(0)
}

func (m *SystemConfigMock) initializeConfigFiles(ctx context.Context, dir string) error {
	args := m.Called(ctx, dir)
	return args.Error(0)
}

func (m *SystemConfigMock) readConfigFile(configFilePath string) ([]byte, error) {
	args := m.Called(configFilePath)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *SystemConfigMock) writeConfigFile(configFilePath string, configFileBytes []byte) error {
	args := m.Called(configFilePath, configFileBytes)
	return args.Error(0)
}

func (m *SystemConfigMock) lock() {
	m.Called()
}

func (m *SystemConfigMock) unlock() {
	m.Called()
}
