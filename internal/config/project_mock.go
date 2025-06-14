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

	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/stretchr/testify/mock"
)

type ProjectConfigMock struct {
	mock.Mock
}

func NewProjectConfigMock() *ProjectConfigMock {
	return &ProjectConfigMock{}
}

func (m *ProjectConfigMock) AddDefaultMocks() {
	m.On("GetManifestSource", mock.Anything).Return(ManifestSourceLocal, nil)
}

func (m *ProjectConfigMock) InitProjectID(ctx context.Context, overwriteExistingProjectID bool) (string, error) {
	args := m.Called(ctx, overwriteExistingProjectID)
	return args.String(0), args.Error(1)
}

func (m *ProjectConfigMock) GetProjectID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *ProjectConfigMock) SetProjectID(ctx context.Context, projectID string) (string, error) {
	args := m.Called(ctx, projectID)
	return args.String(0), args.Error(1)
}

func (m *ProjectConfigMock) GetManifestSource(ctx context.Context) (ManifestSource, error) {
	args := m.Called(ctx)
	return args.Get(0).(ManifestSource), args.Error(1)
}

func (m *ProjectConfigMock) SetManifestSource(ctx context.Context, source ManifestSource) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *ProjectConfigMock) GetSurveyConfig(ctx context.Context, id string) (SurveyConfig, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(SurveyConfig), args.Error(1)
}

func (m *ProjectConfigMock) SetSurveyConfig(ctx context.Context, id string, surveyConfig SurveyConfig) error {
	args := m.Called(ctx, id, surveyConfig)
	return args.Error(0)
}

func (m *ProjectConfigMock) ReadProjectConfigFile(ctx context.Context) (ProjectConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).(ProjectConfig), args.Error(1)
}

func (m *ProjectConfigMock) WriteProjectConfigFile(ctx context.Context, projectConfig ProjectConfig) (string, error) {
	args := m.Called(ctx, projectConfig)
	return args.String(0), args.Error(1)
}

// Cache returns a persistent mock cache
func (m *ProjectConfigMock) Cache() cache.Cacher {
	args := m.Called()
	return args.Get(0).(*cache.CacheMock)
}
