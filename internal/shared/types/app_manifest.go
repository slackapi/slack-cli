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

	"gopkg.in/yaml.v2"
)

const SchemaVersion = 1

// omitempty should be consistent across yaml and json marshalling
// for yaml, are we always using flow style? most properties have it, but when

type AppManifest struct {
	Metadata           *ManifestMetadata  `json:"_metadata,omitempty" yaml:"_metadata,flow,omitempty"`
	DisplayInformation DisplayInformation `json:"display_information" yaml:"display_information,flow"`
	Directory          *AppDirectory      `json:"app_directory,omitempty" yaml:"app_directory,omitempty"`
	Features           *AppFeatures       `json:"features,omitempty" yaml:"features,omitempty"`
	OAuthConfig        *OAuthConfig       `json:"oauth_config,omitempty" yaml:"oauth_config,omitempty"`
	Settings           *AppSettings       `json:"settings,omitempty" yaml:"settings,omitempty"`

	Functions             map[string]ManifestFunction  `json:"functions,omitempty" yaml:"functions,flow,omitempty"`
	Datastores            map[string]ManifestDatastore `json:"datastores,omitempty" yaml:"datastores,flow,omitempty"`
	Types                 *RawJSON                     `json:"types,omitempty" yaml:"types,flow,omitempty"`
	Events                *RawJSON                     `json:"events,omitempty" yaml:"events,flow,omitempty"`
	TriggerTypes          *RawJSON                     `json:"trigger_types,omitempty" yaml:"trigger_types,flow,omitempty"`
	Workflows             map[string]Workflow          `json:"workflows,omitempty" yaml:"workflows,flow,omitempty"`
	OutgoingDomains       *[]string                    `json:"outgoing_domains,omitempty" yaml:"outgoing_domains,flow,omitempty"`
	ExternalAuthProviders *ManifestAuthProviders       `json:"external_auth_providers,omitempty" yaml:"external_auth_providers,flow,omitempty"`
}

type ManifestMetadata struct {
	MajorVersion uint64 `json:"major_version,omitempty" yaml:"major_version,omitempty"`
	MinorVersion uint64 `json:"minor_version,omitempty" yaml:"minor_version,omitempty"`
}

type AppDirectory struct {
	Categories              []string `json:"app_directory_categories,omitempty" yaml:"app_directory_categories,flow,omitempty"`
	UseDirectInstall        *bool    `json:"use_direct_install,omitempty" yaml:"use_direct_install,omitempty"`
	DirectInstallURL        string   `json:"direct_install_url,omitempty" yaml:"direct_install_url,omitempty"`
	InstallationLandingPage string   `json:"installation_landing_page" yaml:"installation_landing_page"`
	PrivacyPolicyURL        string   `json:"privacy_policy_url" yaml:"privacy_policy_url"`
	SupportURL              string   `json:"support_url" yaml:"support_url"`
	SupportEmail            string   `json:"support_email" yaml:"support_email"`
	SupportedLanguages      []string `json:"supported_languages" yaml:"supported_languages,flow"`
	Pricing                 string   `json:"pricing" yaml:"pricing"`
}

type DisplayInformation struct {
	Name            string `json:"name" yaml:"name"`
	Description     string `json:"description,omitempty" yaml:"description,omitempty"`
	BackgroundColor string `json:"background_color,omitempty" yaml:"background_color,omitempty"`
	LongDescription string `json:"long_description,omitempty" yaml:"long_description,omitempty"`
}

type AppFeatures struct {
	AppHome                    ManifestAppHome             `json:"app_home,omitempty" yaml:"app_home,flow,omitempty"`
	AssistantView              *AssistantView              `json:"assistant_view,omitempty" yaml:"assistant_view,omitempty"`
	BotUser                    BotUser                     `json:"bot_user,omitempty" yaml:"bot_user,flow,omitempty"`
	WorkflowSteps              []WorkflowStep              `json:"workflow_steps,omitempty" yaml:"workflow_steps,flow,omitempty"`
	UnfurlDomains              []string                    `json:"unfurl_domains,omitempty" yaml:"unfurl_domains,flow,omitempty"`
	ManifestShortcutsItems     []ManifestShortcutsItem     `json:"shortcuts,omitempty" yaml:"shortcuts,flow,omitempty"`
	ManifestSlashCommandsItems []ManifestSlashCommandsItem `json:"slash_commands,omitempty" yaml:"slash_commands,flow,omitempty"`
}

type AssistantView struct {
	AssistantDescription string             `json:"assistant_description,omitempty" yaml:"assistant_description,omitempty"`
	SuggestedPrompts     []SuggestedPrompts `json:"suggested_prompts,omitempty" yaml:"suggested_prompts,flow,omitempty"`
}

type SuggestedPrompts struct {
	Title   string `json:"title,omitempty" yaml:"title,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type OAuthConfig struct {
	RedirectURLs           []string        `json:"redirect_urls,omitempty" yaml:"redirect_urls,flow,omitempty"`
	Scopes                 *ManifestScopes `json:"scopes,omitempty" yaml:"scopes,flow,omitempty"`
	TokenManagementEnabled *bool           `json:"token_management_enabled,omitempty" yaml:"token_management_enabled,omitempty"`
}

type AppSettings struct {
	SocketModeEnabled      *bool                       `json:"socket_mode_enabled,omitempty" yaml:"socket_mode_enabled,omitempty"`
	OrgDeployEnabled       *bool                       `json:"org_deploy_enabled,omitempty" yaml:"org_deploy_enabled,omitempty"`
	Interactivity          *ManifestInteractivity      `json:"interactivity,omitempty" yaml:"interactivity,omitempty"`
	IncomingWebhooks       *IncomingWebhooks           `json:"incoming_webhooks,omitempty" yaml:"incoming_webhooks,flow,omitempty"`
	EventSubscriptions     *ManifestEventSubscriptions `json:"event_subscriptions,omitempty" yaml:"event_subscriptions,flow,omitempty"`
	AllowedIPAddressRanges []string                    `json:"allowed_ip_address_ranges,omitempty" yaml:"allowed_ip_address_ranges,flow,omitempty"`
	FunctionRuntime        FunctionRuntime             `json:"function_runtime,omitempty" yaml:"function_runtime,flow,omitempty"`
	TokenRotationEnabled   *bool                       `json:"token_rotation_enabled,omitempty" yaml:"token_rotation_enabled,omitempty"`
	SiwsLinks              *SiwsLinks                  `json:"siws_links,omitempty" yaml:"siws_links,flow,omitempty"`
}

type WorkflowStep struct {
	Name       string `json:"name" yaml:"name"`
	CallbackID string `json:"callback_id" yaml:"callback_id"`
}

type BotUser struct {
	DisplayName  string `json:"display_name" yaml:"display_name"`
	AlwaysOnline *bool  `json:"always_online,omitempty" yaml:"always_online,omitempty"`
}

type IncomingWebhooks struct {
	IsEnabled *bool `json:"incoming_webhooks_enabled,omitempty" yaml:"incoming_webhooks_enabled,omitempty"`
}

// ManifestDatastore defines the structure of a datastore in the app manifest.
type ManifestDatastore struct {
	PrimaryKey          string                       `json:"primary_key,omitempty" yaml:"primary_key,omitempty"`
	TimeToLiveAttribute string                       `json:"time_to_live_attribute,omitempty" yaml:"time_to_live_attribute,omitempty"`
	Attributes          map[string]ManifestAttribute `json:"attributes,omitempty" yaml:"attributes,flow,omitempty"`
}

// ManifestAttribute defines the structure of a datastore attribute in the app
// manifest.
type ManifestAttribute struct {
	Type        string   `json:"type,omitempty" yaml:"type,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Items       *RawJSON `json:"items,omitempty" yaml:"items,flow,omitempty"`
	Properties  *RawJSON `json:"properties,omitempty" yaml:"properties,flow,omitempty"`
}

// ManifestFunction defines the structure of a function in the app manifest.
type ManifestFunction struct {
	Title            string   `json:"title" yaml:"title"`
	Description      string   `json:"description" yaml:"description"`
	Type             string   `json:"type,omitempty" yaml:"type,omitempty"`
	Bindings         *RawJSON `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	InputParameters  *RawJSON `json:"input_parameters" yaml:"input_parameters,flow"`
	OutputParameters *RawJSON `json:"output_parameters" yaml:"output_parameters,flow"`
}

type ManifestAuthProviders struct {
	OAuth2 map[string]*RawJSON `json:"oauth2" yaml:"oauth2"`
}

type RawJSON struct {
	Data     *yaml.MapSlice
	JSONData *json.RawMessage
}

// TODO (@kattari) Should we remove this? It's not used anywhere.
func (r *RawJSON) MarshalJSON() ([]byte, error) {
	if r.JSONData != nil {
		return *(r.JSONData), nil
	}
	var d = r.convertData(*((*r).Data))
	return json.Marshal(d)
}

func (r *RawJSON) UnmarshalJSON(data []byte) error {
	var rawJSON json.RawMessage = data
	r.JSONData = &rawJSON
	return nil
}

func (r *RawJSON) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&r.Data)
}

// convertData is a recursive function that takes any object and converts all
// instances of yaml.MapSlice (nested or not) into a map[string]interface.
// For any other type, it will return it exactly as is.
// For example:
//
//	[{Key: "name", Value: "foo"}, {Key: "age", Value: "1"}] will turn into
//	map[string]interface{}{"name": "foo", "age": "1"}
func (r *RawJSON) convertData(val interface{}) interface{} {
	switch val := val.(type) {
	case yaml.MapSlice:
		// convert the MapSlice into a proper golang map[string]interface{}{}
		var res = map[string]interface{}{}
		for _, mi := range val {
			res[mi.Key.(string)] = r.convertData(mi.Value)
		}
		return res
	case []interface{}:
		// if the value is rather a slice of any type, we cannot assume
		// that there are no yaml.MapSlice values in there so we have to
		// recursively convert all the contents.
		var res []interface{}
		for _, v := range val {
			res = append(res, r.convertData(v))
		}
		return res
	default: // base case
		return val
	}
}

// ManifestAppHome
type ManifestAppHome struct {
	HomeTabEnabled             *bool `json:"home_tab_enabled,omitempty" yaml:"home_tab_enabled,omitempty"`
	MessagesTabEnabled         *bool `json:"messages_tab_enabled,omitempty" yaml:"messages_tab_enabled,omitempty"`
	MessagesTabReadOnlyEnabled *bool `json:"messages_tab_read_only_enabled,omitempty" yaml:"messages_tab_read_only_enabled,omitempty"`
}

// ManifestEventSubscriptions
type ManifestEventSubscriptions struct {
	RequestURL            string                 `json:"request_url,omitempty" yaml:"request_url,omitempty"`
	UserEvents            []string               `json:"user_events,omitempty" yaml:"user_events,flow,omitempty"`
	BotEvents             []string               `json:"bot_events,omitempty" yaml:"bot_events,flow,omitempty"`
	MetadataSubscriptions []MetadataSubscription `json:"metadata_subscriptions,omitempty" yaml:"metadata_subscriptions,flow,omitempty"`
}

type MetadataSubscription struct {
	AppID     string `json:"app_id" yaml:"app_id"`
	EventType string `json:"event_type" yaml:"event_type"`
}

type ManifestInteractivity struct {
	IsEnabled             bool   `json:"is_enabled" yaml:"is_enabled"`
	RequestURL            string `json:"request_url,omitempty" yaml:"request_url,omitempty"`
	MessageMenuOptionsURL string `json:"message_menu_options_url,omitempty" yaml:"message_menu_options_url,omitempty"`
}

// ManifestScopes
type ManifestScopes struct {
	Bot  []string `json:"bot,omitempty" yaml:"bot,flow,omitempty"`
	User []string `json:"user,omitempty" yaml:"user,flow,omitempty"`
}

// ManifestShortcutsItem
type ManifestShortcutsItem struct {
	CallbackID  string            `json:"callback_id" yaml:"callback_id"`
	Description string            `json:"description" yaml:"description"`
	Name        string            `json:"name" yaml:"name"`
	Type        ShortcutScopeType `json:"type" yaml:"type"`
}

// ManifestSlashCommandsItem
type ManifestSlashCommandsItem struct {
	Command      string `json:"command" yaml:"command"`
	URL          string `json:"url,omitempty" yaml:"url,omitempty"`
	Description  string `json:"description" yaml:"description"`
	ShouldEscape *bool  `json:"should_escape,omitempty" yaml:"should_escape,omitempty"`
	UsageHint    string `json:"usage_hint,omitempty" yaml:"usage_hint,omitempty"`
}

// Workflow defines the structure of a workflow in the app manifest.
type Workflow struct {
	Title             string             `json:"title" yaml:"title"`
	Description       string             `json:"description" yaml:"description"`
	InputParameters   *RawJSON           `json:"input_parameters,omitempty" yaml:"input_parameters,flow,omitempty"`
	Steps             []Step             `json:"steps" yaml:"steps,flow"`
	SuggestedTriggers []SuggestedTrigger `json:"suggested_triggers,omitempty" yaml:"suggested_triggers,flow,omitempty"`
}

// Step defines the structure of a step in the app manifest.
type Step struct {
	ID         string   `json:"id" yaml:"id"`
	FunctionID string   `json:"function_id" yaml:"function_id"`
	Inputs     *RawJSON `json:"inputs" yaml:"inputs,flow"`
}

// SuggestedTrigger defines the structure of a suggested trigger in the app
// manifest.
type SuggestedTrigger struct {
	Type        string            `json:"type" yaml:"type"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Inputs      map[string]string `json:"inputs" yaml:"inputs,flow"`
}

// Constraint defines the structure of a constraint in the app manifest.
type Constraint struct {
	Type  string `json:"type" yaml:"type"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

type SiwsLinks struct {
	InitiateURI string `json:"initiate_uri,omitempty" yaml:"initiate_uri,omitempty"`
}

type FunctionRuntime string

const (
	LocallyRun  FunctionRuntime = "local"
	Remote      FunctionRuntime = "remote"
	SlackHosted FunctionRuntime = "slack"
)

type ShortcutScopeType string

// Methods

// FunctionRuntime returns the FunctionRuntime of an app manifest if exists
func (manifest *AppManifest) FunctionRuntime() FunctionRuntime {
	if manifest == nil || manifest.Settings == nil {
		return ""
	}
	return manifest.Settings.FunctionRuntime
}

// IsFunctionRuntimeSlackHosted returns true when the function runtime setting
// is slack hosted
func (manifest *AppManifest) IsFunctionRuntimeSlackHosted() bool {
	return manifest.Settings != nil && manifest.Settings.FunctionRuntime == SlackHosted
}

// ToRawJSON converts a string to types.RawJSON
func ToRawJSON(obj string) *RawJSON {
	b := []byte(obj)
	return &RawJSON{JSONData: (*json.RawMessage)(&b)}
}
