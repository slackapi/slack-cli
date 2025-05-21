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

package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func Test_RawJSON_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name              string
		blob              string
		expectedErrorType error
		expectedJSONData  string
	}{
		{
			name:              "Unmarshal data",
			blob:              `{ "name": "foo" }`,
			expectedErrorType: nil,
			expectedJSONData:  `{ "name": "foo" }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawJSON := &RawJSON{}
			err := rawJSON.UnmarshalJSON([]byte(tt.blob))

			require.IsType(t, err, tt.expectedErrorType)
			require.Equal(t, tt.expectedJSONData, string(*rawJSON.JSONData))
		})
	}
}

func Test_RawJSON_UnmarshalYAML(t *testing.T) {
	rawJSON := RawJSON{Data: &yaml.MapSlice{
		{Key: "name", Value: "foo"},
	}}

	yamlUnmarshaler := &YAMLUnmarshalerMock{}
	yamlUnmarshaler.On("unmarshal", mock.Anything).Return(nil)

	err := rawJSON.UnmarshalYAML(yamlUnmarshaler.unmarshal)

	yamlUnmarshaler.AssertCalled(t, "unmarshal", &rawJSON.Data)
	assert.NoError(t, err)
}

func Test_AppManifest_ConvertDataForRawJSON(t *testing.T) {
	tests := map[string]struct {
		have RawJSON
		want interface{}
	}{
		"basic": {
			have: RawJSON{Data: &yaml.MapSlice{
				yaml.MapItem{Key: "name", Value: "foo"},
			}},
			want: map[string]string{"name": "foo"},
		},
		"nested": {
			have: RawJSON{Data: &yaml.MapSlice{
				{Key: "name", Value: "foo"},
				{Key: "about", Value: yaml.MapSlice{{Key: "age", Value: "1"}}},
			}},
			want: map[string]interface{}{"name": "foo", "about": map[string]string{"age": "1"}},
		},
		"include_slices": {
			have: RawJSON{Data: &yaml.MapSlice{
				{Key: "title", Value: "Title"},
				{Key: "fruits", Value: []string{"mango", "pineapple"}},
				{Key: "vegetables", Value: []string{"onion", "ginger"}},
			}},
			want: map[string]interface{}{"title": "Title", "fruits": []string{"mango", "pineapple"}, "vegetables": []string{"onion", "ginger"}},
		},
		"interface slices": {
			have: RawJSON{Data: &yaml.MapSlice{
				yaml.MapItem{Key: "name", Value: "foo"},
				yaml.MapItem{Key: "fruits", Value: []interface{}{"mango", "pineapple"}},
			}},
			want: map[string]interface{}{"name": "foo", "fruits": []string{"mango", "pineapple"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			have, err := json.Marshal(tt.have.convertData(*tt.have.Data))
			assert.Nil(err)
			want, err := json.Marshal(tt.want)
			assert.Nil(err)
			assert.Equal(want, have)
		})
	}
}

func Test_AppManifest_ToRawJSON(t *testing.T) {
	tests := map[string]struct {
		have string
		want *RawJSON
	}{
		"empty json object": {
			have: "",
			want: &RawJSON{JSONData: &json.RawMessage{}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := ToRawJSON(tt.have); !reflect.DeepEqual(got, tt.want) {
				t.Log(got.Data)
				t.Log(got.JSONData)
				t.Errorf("ToRawJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_AppManifest_AppFeatures(t *testing.T) {
	truth := true
	tests := map[string]struct {
		features AppFeatures
		want     string
	}{
		"includes provided values without blank defaults": {
			features: AppFeatures{
				AppHome: ManifestAppHome{
					HomeTabEnabled:     &truth,
					MessagesTabEnabled: &truth,
				},
				BotUser: BotUser{
					DisplayName: "slackbot",
				},
			},
			want: `{"app_home":{"home_tab_enabled":true,"messages_tab_enabled":true},"bot_user":{"display_name":"slackbot"}}`,
		},
		"includes assistant view when provided": {
			features: AppFeatures{
				AssistantView: &AssistantView{
					AssistantDescription: "magic",
					SuggestedPrompts: []SuggestedPrompts{
						{
							Title:   "visit the beach",
							Message: "what is glass",
						},
					},
				},
				BotUser: BotUser{
					DisplayName: "einstein",
				},
			},
			want: `{"app_home":{},"assistant_view":{"assistant_description":"magic","suggested_prompts":[{"title":"visit the beach","message":"what is glass"}]},"bot_user":{"display_name":"einstein"}}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := json.Marshal(tt.features)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(actual))
		})
	}
}

func Test_AppManifest_AppSettings_SiwsLinks(t *testing.T) {
	expectedSiws := SiwsLinks{
		InitiateURI: "an initiate uri",
	}
	tests := map[string]struct {
		settings          *AppSettings
		expectedSiwsLinks *SiwsLinks
		expectedJSON      string
	}{
		"undefined incoming webhooks have no siws links": {
			settings:          &AppSettings{},
			expectedSiwsLinks: nil,
			expectedJSON:      `{}`,
		},
		"defined siws links have siws links": {
			settings:          &AppSettings{SiwsLinks: &expectedSiws},
			expectedSiwsLinks: &expectedSiws,
			expectedJSON:      `{"siws_links":{"initiate_uri":"an initiate uri"}}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := AppManifest{
				Settings: tt.settings,
			}
			if tt.settings != nil {
				actualJSON, err := json.Marshal(tt.settings)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedJSON, string(actualJSON))
				assert.Equal(t, tt.expectedSiwsLinks, manifest.Settings.SiwsLinks)
			} else {
				assert.Nil(t, manifest.Settings)
			}
		})
	}
}

func Test_AppManifest_AppSettings_IncomingWebhooks(t *testing.T) {
	falsity := false
	expectedIncomingWebhooks := IncomingWebhooks{
		IsEnabled: &falsity,
	}
	tests := map[string]struct {
		settings                      *AppSettings
		expectedIncomingWebhooksLinks *IncomingWebhooks
		expectedJSON                  string
	}{
		"undefined incoming webhooks have no webhooks": {
			settings:                      &AppSettings{},
			expectedIncomingWebhooksLinks: nil,
			expectedJSON:                  `{}`,
		},
		"defined incoming webhooks have webhooks": {
			settings:                      &AppSettings{IncomingWebhooks: &expectedIncomingWebhooks},
			expectedIncomingWebhooksLinks: &expectedIncomingWebhooks,
			expectedJSON:                  `{"incoming_webhooks":{"incoming_webhooks_enabled":false}}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := AppManifest{
				Settings: tt.settings,
			}
			if tt.settings != nil {
				actualJSON, err := json.Marshal(tt.settings)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedJSON, string(actualJSON))
				assert.Equal(t, tt.expectedIncomingWebhooksLinks, manifest.Settings.IncomingWebhooks)
			} else {
				assert.Nil(t, manifest.Settings)
			}
		})
	}
}

func Test_AppManifest_AppSettings_FunctionRuntime(t *testing.T) {
	tests := map[string]struct {
		settings        *AppSettings
		expectedHosted  bool
		expectedRuntime FunctionRuntime
	}{
		"undefined settings have no function runtime": {
			settings:        nil,
			expectedHosted:  false,
			expectedRuntime: "",
		},
		"undefined function runtime has no function runtime": {
			settings:        &AppSettings{},
			expectedHosted:  false,
			expectedRuntime: "",
		},
		"setting the function runtime to slack is hosted": {
			settings:        &AppSettings{FunctionRuntime: "slack"},
			expectedHosted:  true,
			expectedRuntime: SlackHosted,
		},
		"setting the function runtime to remote is not hosted": {
			settings:        &AppSettings{FunctionRuntime: "remote"},
			expectedHosted:  false,
			expectedRuntime: Remote,
		},
		"setting the function runtime to local is not hosted": {
			settings:        &AppSettings{FunctionRuntime: "local"},
			expectedHosted:  false,
			expectedRuntime: LocallyRun,
		},
		"setting the function runtime to random is not hosted": {
			settings:        &AppSettings{FunctionRuntime: "sparkling-butterflies"},
			expectedHosted:  false,
			expectedRuntime: "sparkling-butterflies",
		},
		"setting the function runtime to padded string is possible": {
			settings:        &AppSettings{FunctionRuntime: "    "},
			expectedHosted:  false,
			expectedRuntime: "    ",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := AppManifest{
				Settings: tt.settings,
			}
			assert.Equal(t, tt.expectedHosted, manifest.IsFunctionRuntimeSlackHosted())
			assert.Equal(t, tt.expectedRuntime, manifest.FunctionRuntime())
			if tt.settings != nil {
				assert.Equal(t, tt.expectedRuntime, manifest.Settings.FunctionRuntime)
			}
		})
	}
}
