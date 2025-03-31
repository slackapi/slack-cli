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

package iostreams

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MatchPromptConfig helps find the prompt with a matching config for mocking
func MatchPromptConfig(match PromptConfig) interface{} {
	return mock.MatchedBy(func(config PromptConfig) bool {
		matchFlags := match.GetFlags()
		configFlags := config.GetFlags()
		if len(matchFlags) != len(configFlags) {
			return false
		}
		for _, mFlag := range matchFlags {
			match := false
			for _, cFlag := range configFlags {
				if cFlag.Name == mFlag.Name {
					match = true
				}
			}
			if !match {
				return false
			}
		}
		return true
	})
}

// ConfirmPrompt mocks the confirm prompt from go-survey
func (m *IOStreamsMock) ConfirmPrompt(ctx context.Context, message string, defaultValue bool) (bool, error) {
	args := m.Called(ctx, message, defaultValue)
	return args.Bool(0), nil
}

// InputPrompt mocks the input prompt from go-survey
func (m *IOStreamsMock) InputPrompt(ctx context.Context, message string, cfg InputPromptConfig) (string, error) {
	args := m.Called(ctx, message, cfg)
	return args.String(0), nil
}

// SelectPrompt mocks the select prompt from go-survey
func (m *IOStreamsMock) MultiSelectPrompt(ctx context.Context, message string, options []string) ([]string, error) {
	args := m.Called(ctx, message)
	return args.Get(0).([]string), nil
}

// PasswordPrompt mocks the password prompt or flag selection
func (m *IOStreamsMock) PasswordPrompt(ctx context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error) {
	args := m.Called(ctx, message, cfg)
	return args.Get(0).(PasswordPromptResponse), args.Error(1)
}

// SelectPrompt mocks the select prompt from go-survey
func (m *IOStreamsMock) SelectPrompt(ctx context.Context, message string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error) {
	args := m.Called(ctx, message, options, cfg)
	return args.Get(0).(SelectPromptResponse), args.Error(1)
}
