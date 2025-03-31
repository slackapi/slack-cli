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
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

// NewCobraMock creates a mock for some Cobra functions.
func NewCobraMock() *CobraMock {
	return &CobraMock{}
}

// CobraMock mocks the client's Cobra struct.
type CobraMock struct {
	mock.Mock
}

// AddDefaultMocks installs the default mock actions to fallback on.
func (m *CobraMock) AddDefaultMocks() {
	m.On("GenMarkdownTree", mock.Anything, mock.Anything).Return(nil)
}

// GenMarkdownTree mock.
func (m *CobraMock) GenMarkdownTree(cmd *cobra.Command, dir string) error {
	args := m.Called(cmd, dir)
	return args.Error(0)
}
