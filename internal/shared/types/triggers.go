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

import "encoding/json"

// Supported trigger types
const (
	TriggerTypeShortcut        = "shortcut"
	TriggerTypeSlashCommand    = "slash_command"
	TriggerTypeMessageShortcut = "message_shortcut"
	TriggerTypeEvent           = "event"
	TriggerTypeWebhook         = "webhook"
	TriggerTypeScheduled       = "scheduled"
)

// Trigger type is accepted by the `functions.triggers.*` methods
type Trigger struct {
	Type              string          `json:"type" yaml:"type"`                             // all
	FunctionAppID     string          `json:"function_app_id" yaml:"function_app_id"`       // all
	FunctionRef       string          `json:"function_reference" yaml:"function_reference"` // all
	Name              *string         `json:"name,omitempty" yaml:"name"`                   // shortcut, slash_command, message_shortcut
	Description       *string         `json:"description,omitempty" yaml:"description"`     // shortcut, slash_command, message_shortcut
	EventType         *string         `json:"event_type,omitempty"`                         // event
	MetadataEventType *string         `json:"metadata_event_type,omitempty"`                // event
	Schedule          *RawJSON        `json:"schedule,omitempty" yaml:"schedule"`           // scheduled
	Inputs            *RawJSON        `json:"inputs" yaml:"inputs,flow"`                    // all
	Filter            *RawJSON        `json:"filter,omitempty" yaml:"filter,flow"`          // all
	ChannelIDs        []string        `json:"channel_ids,omitempty" yaml:"channel_ids"`     // all
	RawMessage        json.RawMessage `json:"-" yaml:"-"`                                   // "-" will always omit
}

// IsKnownType returns true if Trigger.Type is a known and supported type
func (t *Trigger) IsKnownType() bool {
	switch t.Type {
	case
		TriggerTypeShortcut,
		TriggerTypeSlashCommand,
		TriggerTypeMessageShortcut,
		TriggerTypeEvent,
		TriggerTypeWebhook,
		TriggerTypeScheduled:
		return true
	default:
		return false
	}
}

// TriggerDefinition type is returned by the SDK hook 'triggers'
type TriggerDefinition struct {
	Key     string        `json:"key" yaml:"key"`
	Trigger *Trigger      `json:"trigger,omitempty" yaml:"trigger,omitempty"`
	Access  TriggerAccess `json:"access,omitempty" yaml:"access,omitempty"`
}

// RawTriggerDefinition type stores the trigger as raw JSON to handle untyped triggers
type RawTriggerDefinition struct {
	Key        string          `json:"key" yaml:"key"`
	TriggerRaw json.RawMessage `json:"trigger" yaml:"trigger"`
	Access     TriggerAccess   `json:"access,omitempty" yaml:"access,omitempty"`
}

// TriggerMutations type is populated by the CLI when calling the app hook 'triggers'
type TriggerMutations struct {
	Create []TriggerCreateMutation
	Update []TriggerUpdateMutation
	Delete []TriggerDeleteMutation
}

// TriggerCreateMutation type is used by the CLI when calling the app hook 'triggers'
type TriggerCreateMutation struct {
	Error   error
	Trigger TriggerDefinition
}

// TriggerUpdateMutation type is used by the CLI when calling the app hook 'triggers'
type TriggerUpdateMutation struct {
	Error     error
	TriggerID string
	Trigger   TriggerDefinition
}

// TriggerDeleteMutation type is used by the CLI when calling the app hook 'triggers'
type TriggerDeleteMutation struct {
	Error      error
	TriggerKey string
	TriggerID  string
}

// TriggerAccess type describes the access for a particular trigger
type TriggerAccess struct {
	UserIDs []string `json:"user_ids,omitempty" yaml:"user_ids,omitempty"`
}

type TriggerSyncOptions struct {
	App        App
	DeleteAll  bool
	Manifest   AppManifest
	AuthTokens string
}

type TriggerWorkflow struct {
	ID          string `json:"id"`
	CallbackID  string `json:"callback_id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	AppID       string `json:"app_id"`
	App         struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Icons Icons  `json:"icons"`
	}
	InputParameters  *RawJSON `json:"input_parameters"`
	OutputParameters *RawJSON `json:"output_parameters"`
}

type DeployedTrigger struct {
	ID          string          `json:"id"`
	DateCreated int             `json:"date_created"`
	DateUpdated int             `json:"date_updated"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Usage       string          `json:"usage"`
	Webhook     string          `json:"webhook_url"`
	ShortcutURL string          `json:"shortcut_url"`
	Workflow    TriggerWorkflow `json:"workflow"`
	Inputs      *RawJSON        `json:"inputs"`
}
