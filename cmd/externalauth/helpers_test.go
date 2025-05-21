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

package externalauth

import (
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared/types"
)

var fakeAppID = "A1234"
var fakeAppTeamID = "T1234"
var fakeAppUserID = "U1234"

var fakeApp = types.App{
	TeamDomain: "test",
	AppID:      fakeAppID,
	TeamID:     fakeAppTeamID,
	UserID:     fakeAppUserID,
}

var (
	installedProdApp = prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{AppID: fakeAppID}}
)

func setupMockAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = appSelectPromptFunc
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		appSelectPromptFunc = originalPromptFunc
	}
}

func setupMockProviderSelection() func() {
	externalAuthMock := NewExternalAuthMock()
	var originalPromptFunc = providerSelectionFunc
	providerSelectionFunc = externalAuthMock.ProviderSelectPrompt
	externalAuthMock.On("ProviderSelectPrompt").Return(types.ExternalAuthorizationInfo{
		ProviderKey: "provider_a",
	}, nil)
	return func() {
		providerSelectionFunc = originalPromptFunc
	}
}

func setupMockTokenSelection() func() {
	externalAuthMock := NewExternalAuthMock()
	var originalPromptFunc = tokenSelectionFunc
	tokenSelectionFunc = externalAuthMock.TokenSelectPrompt
	externalAuthMock.On("TokenSelectPrompt").Return(types.ExternalTokenInfo{
		ExternalTokenID: "Et0548LABCDE",
	}, nil)
	return func() {
		tokenSelectionFunc = originalPromptFunc
	}
}
