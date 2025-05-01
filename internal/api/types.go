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

package api

// Interface that embeds all of the sub-interfaces for each file in this package.
// api.Client implements this interface, so it can be mocked out easily.
// TODO: consider renaming interfaces from '*Client' to an -er verb such as '*Manager'
//
//	Ref: https://go.dev/doc/effective_go#interface-names
type APIInterface interface {
	ActivityClient
	AppsClient
	AuthClient
	ChannelClient
	CollaboratorsClient
	DatastoresClient
	ExternalAuthClient
	FunctionDistributionClient
	SessionsClient
	StepsClient
	TeamClient
	TriggerAccessClient
	UserClient
	VariablesClient
	WorkflowsClient
}
