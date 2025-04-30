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

package shared

import (
	"bytes"
	"strings"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/auth"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/tracking"
	"github.com/stretchr/testify/mock"
)

// ClientsMock defines mocks that will override aspects of clients for testing purposes.
type ClientsMock struct {
	mock.Mock
	APIInterface  *api.APIMock
	AuthInterface *auth.AuthMock
	AppClient     *app.Client
	Browser       *slackdeps.BrowserMock
	Config        *config.Config
	Cobra         *slackdeps.CobraMock
	EventTracker  *tracking.EventTrackerMock
	Fs            *slackdeps.FsMock
	IO            *iostreams.IOStreamsMock
	Os            *slackdeps.OsMock
	Stdout        *bytes.Buffer
	HookExecutor  hooks.MockHookExecutor
}

// NewClientsMock will create a new ClientsMock that is ready to be applied to an existing clients with .MockClientFactory().
func NewClientsMock() *ClientsMock {
	// Create a new clients mock
	clientsMock := &ClientsMock{}

	// Set the mocked members
	clientsMock.APIInterface = &api.APIMock{}
	clientsMock.AuthInterface = &auth.AuthMock{}
	clientsMock.Browser = slackdeps.NewBrowserMock()
	clientsMock.Cobra = slackdeps.NewCobraMock()
	clientsMock.EventTracker = &tracking.EventTrackerMock{}
	clientsMock.Fs = slackdeps.NewFsMock()
	clientsMock.Os = slackdeps.NewOsMock()
	clientsMock.HookExecutor = hooks.MockHookExecutor{}

	clientsMock.Config = config.NewConfig(clientsMock.Fs, clientsMock.Os)
	clientsMock.IO = iostreams.NewIOStreamsMock(clientsMock.Config, clientsMock.Fs, clientsMock.Os)

	clientsMock.AppClient = &app.Client{Manifest: &app.ManifestMockObject{}, AppClientInterface: app.NewAppClient(clientsMock.Config, clientsMock.Fs, clientsMock.Os)}

	return clientsMock
}

// AddDefaultMocks installs the default mock actions to fallback on.
func (m *ClientsMock) AddDefaultMocks() {
	m.APIInterface.AddDefaultMocks()
	m.AuthInterface.AddDefaultMocks()
	m.Browser.AddDefaultMocks()
	m.Cobra.AddDefaultMocks()
	m.EventTracker.AddDefaultMocks()
	m.IO.AddDefaultMocks()
	m.Os.AddDefaultMocks()
}

// MockClientFactory is a functional option that installs mocks into a clients instance.
// Learn more: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func (m *ClientsMock) MockClientFactory() func(c *ClientFactory) {
	return func(clients *ClientFactory) {
		clients.Browser = func() slackdeps.Browser { return m.Browser }
		clients.Cobra.GenMarkdownTree = m.Cobra.GenMarkdownTree
		clients.Config = m.Config
		clients.EventTracker = m.EventTracker
		clients.Os = m.Os
		clients.IO = m.IO
		clients.Fs = m.Fs
		clients.APIInterface = func() api.APIInterface { return m.APIInterface }
		clients.AuthInterface = func() auth.AuthInterface { return m.AuthInterface }
		clients.AppClient = func() *app.Client { return m.AppClient }
		clients.HookExecutor = &m.HookExecutor
	}
}

// GetCombinedOutput is a helper method to return the combined output of stdout and stderr for testing.
func (m *ClientsMock) GetCombinedOutput() string {
	outputs := []string{
		m.GetStdoutOutput(),
		m.GetStderrOutput(),
	}
	return strings.Join(outputs, "\n")
}

// GetStdoutOutput is a helper method to return the stdout output for testing.
func (m *ClientsMock) GetStdoutOutput() string {
	if stdout, ok := m.IO.WriteOut().(*bytes.Buffer); ok {
		return stdout.String()
	}
	return ""
}

// GetStderrOutput is a helper method to return the stderr output for testing.
func (m *ClientsMock) GetStderrOutput() string {
	if stderr, ok := m.IO.WriteErr().(*bytes.Buffer); ok {
		return stderr.String()
	}
	return ""
}
