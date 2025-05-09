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

type ProviderData struct {
	ProviderType string                         `json:"provider_type" yaml:"provider_type,flow"`
	Options      *ManifestOAuth2ProviderOptions `json:"options" yaml:"options,flow"`
}

type ManifestOAuth2ProviderOptions struct {
	AuthorizationURL       string            `json:"authorization_url" yaml:"authorization_url,flow"`
	ClientID               string            `json:"client_id" yaml:"client_id,flow"`
	TokenURL               string            `json:"token_url" yaml:"token_url,flow"`
	Scope                  []string          `json:"scope" yaml:"scope,flow"`
	ProviderName           string            `json:"provider_name" yaml:"provider_name,flow"`
	IdentityConfig         *RawJSON          `json:"identity_config" yaml:"identity_config,flow"`
	AuthorizationURLExtras map[string]string `json:"authorization_url_extras,omitempty" yaml:"authorization_url_extras,omitempty,flow"`
}

type ExternalAuthorizationInfo struct {
	ProviderName       string              `json:"provider_name" yaml:"provider_name,flow"`
	ProviderKey        string              `json:"provider_key" yaml:"provider_key,flow"`
	ClientID           string              `json:"client_id" yaml:"client_id,flow"`
	ClientSecretExists bool                `json:"client_secret_exists" yaml:"client_secret_exists,flow"`
	ValidTokenExists   bool                `json:"valid_token_exists" yaml:"valid_token_exists,flow"`
	ExternalTokenIDs   []string            `json:"external_token_ids,omitempty" yaml:"external_token_ids,omitempty,flow"`
	ExternalTokens     []ExternalTokenInfo `json:"external_tokens,omitempty" yaml:"external_tokens,omitempty,flow"`
}

type WorkflowsInfo struct {
	WorkflowID string          `json:"workflow_id" yaml:"workflow_id,flow"`
	CallbackID string          `json:"callback_id" yaml:"callback_id,flow"`
	Providers  []ProvidersInfo `json:"providers" yaml:"providers,flow"`
}

type ProvidersInfo struct {
	ProviderName string            `json:"provider_name" yaml:"provider_name,flow"`
	ProviderKey  string            `json:"provider_key" yaml:"provider_key,flow"`
	SelectedAuth ExternalTokenInfo `json:"selected_auth,omitempty" yaml:"selected_auth,omitempty,flow"`
}

type ExternalTokenInfo struct {
	ExternalTokenID string `json:"external_token_id" yaml:"external_token_id,flow"`
	ExternalUserID  string `json:"external_user_id" yaml:"external_user_id,flow"`
	DateUpdated     int    `json:"date_updated" yaml:"date_updated,flow"`
}

type ExternalAuthorizationInfoLists struct {
	Authorizations []ExternalAuthorizationInfo `json:"authorizations" yaml:"authorizations,flow"`
	Workflows      []WorkflowsInfo             `json:"workflows,omitempty" yaml:"workflows,omitempty,flow"`
}

type WorkflowExternalAuthorizationInfo struct {
	WorkflowID string                              `json:"workflow_id" yaml:"workflow_id,flow"`
	CallbackID string                              `json:"callback_id" yaml:"callback_id,flow"`
	Providers  []ProviderExternalAuthorizationInfo `json:"providers" yaml:"providers,flow"`
}

type ProviderExternalAuthorizationInfo struct {
	ProviderName string `json:"provider_name" yaml:"provider_name,flow"`
	ProviderKey  string `json:"provider_key" yaml:"provider_key,flow"`
	SelectedAuth string `json:"selected_auth" yaml:"selected_auth,omitempty"`
}

type SelectedAuthInfo struct {
	ExternalTokenID string `json:"external_token_id" yaml:"external_token_id,flow"`
	ExternalUserID  string `json:"external_user_id" yaml:"external_user_id,flow"`
	DateUpdated     string `json:"date_updated" yaml:"date_updated,flow"`
}
