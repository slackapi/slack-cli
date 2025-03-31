// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SlackChannel_String(t *testing.T) {
	tests := []struct {
		name           string
		channel        *SlackChannel
		expectedString string
	}{
		{
			name:           "Both ChannelName and ID exist",
			channel:        &SlackChannel{ChannelName: "#general", ID: "C01234"},
			expectedString: "#general (C01234)",
		},
		// TODO(@mbrooks) This test represents the current behaviour, but should the expectedString be "#general" instead?
		{
			name:           "Only ChannelName exists",
			channel:        &SlackChannel{ChannelName: "#general"},
			expectedString: "(#general)",
		},
		// TODO(@mbrooks) This test represents the current behaviour, but should the expectedString be "C01234" instead?
		{
			name:           "Only ID exists",
			channel:        &SlackChannel{ID: "C01234"},
			expectedString: "",
		},
		{
			name:           "Both ChannelName and ID do not exist",
			channel:        &SlackChannel{},
			expectedString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.channel.String(), tt.expectedString)
		})
	}
}
