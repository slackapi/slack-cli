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

package slackdeps

import (
	"os"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

// MockHomeDirectory is the default returned by UserHomeDir().
const MockHomeDirectory = "/Users/user.name"

// MockWorkingDirectory is the default returned by Getwd().
const MockWorkingDirectory = "/Users/user.name/app"

// MockCustomConfigDirectory is a custom config directory for ConfigDirFlag
const MockCustomConfigDirectory = "/tmp/tmp.user.name.123"

// NewOsMock creates a new OsMock
func NewOsMock() *OsMock {
	return &OsMock{}
}

// OsMock mocks the client's Os struct.
type OsMock struct {
	mock.Mock
}

// AddDefaultMocks installs the default mock actions to fallback on.
func (m *OsMock) AddDefaultMocks() {
	m.On("Getenv", mock.Anything).Return("")
	m.On("LookPath", mock.Anything).Return("", nil)
	m.On("LookupEnv", mock.Anything).Return("", false)
	m.On("Setenv", mock.Anything, mock.Anything).Return(nil)
	m.On("Getwd").Return(MockWorkingDirectory, nil)
	m.On("UserHomeDir").Return(MockHomeDirectory, nil)
	m.On("GetExecutionDir").Return(MockHomeDirectory, nil)
	m.On("SetExecutionDir", mock.Anything)
	m.On("IsNotExist", mock.Anything).Return(true)
	m.On("Exit", mock.Anything).Return()
	m.On("Stat", mock.Anything).Return()
	m.On("Stdout").Return(os.Stdout)
}

// Getenv mocks returning an environment variable
func (m *OsMock) Getenv(key string) (value string) {
	args := m.Called(key)
	return args.String(0)
}

// LookPath mocks finding the path of an executable
func (m *OsMock) LookPath(file string) (path string, err error) {
	args := m.Called(file)
	return args.String(0), args.Error(1)
}

// LookupEnv mocks searching for an environment variable
func (m *OsMock) LookupEnv(key string) (value string, present bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

// Setenv mocks the setting of an environment variable
func (m *OsMock) Setenv(key string, value string) error {
	m.On("Getenv", key).Return(value)
	m.On("LookupEnv", key).Return(value, true)
	args := m.Called(key, value)
	return args.Error(0)
}

// Getwd mocks returning the working directory.
func (m *OsMock) Getwd() (_dir string, _err error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// UserHomeDir mocks returning the home directory.
func (m *OsMock) UserHomeDir() (_dir string, _err error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// GetExecutionDir mocks returning the absolute path where the process started execution
func (m *OsMock) GetExecutionDir() string {
	args := m.Called()
	return args.String(0)
}

// SetExecutionDir mocks setting the absolute path where the process started execution
func (m *OsMock) SetExecutionDir(dirPathAbs string) {
	_ = m.Called(dirPathAbs)
}

// IsNotExist returns a boolean indicating whether the provided error is known to report that a file or directory does not exist
func (m *OsMock) IsNotExist(err error) (_bool bool) {
	args := m.Called(err)
	return args.Bool(0)
}

func (m *OsMock) Glob(pattern string) (matches []string, err error) {
	args := m.Called(pattern)
	return args.Get(0).([]string), args.Error(1)
}

// Exit mocks exiting the program with a return code, but does not actually exit
func (m *OsMock) Exit(code int) {
	m.Called(code)
}

// Stdout mocks the stdout with a file that can be adjusted
func (m *OsMock) Stdout() types.File {
	args := m.Called()
	return args.Get(0).(types.File)
}
