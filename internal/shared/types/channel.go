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

import "fmt"

type SlackChannel struct {
	ID          string `json:"channel_id,omitempty" yaml:"channel_id,omitempty"`
	ChannelName string `json:"channel_name,omitempty" yaml:"channel_name,omitempty"`
}

func (c *SlackChannel) String() string {
	if c.ID != "" && c.ChannelName != "" {
		return fmt.Sprintf("%s (%s)", c.ChannelName, c.ID)
	} else if c.ChannelName != "" {
		return fmt.Sprintf("(%s)", c.ChannelName)
	}
	return c.ChannelName
}

// Channel model with fields that match what is returned from the conversations.info method
// Method documentation: https://docs.slack.dev/reference/methods/conversations.info
type ChannelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
