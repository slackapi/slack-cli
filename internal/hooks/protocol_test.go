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

package hooks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Protocol_String(t *testing.T) {
	var p Protocol

	p = HookProtocolDefault
	require.Equal(t, string(HookProtocolDefault), p.String())

	p = HookProtocolV2
	require.Equal(t, string(HookProtocolV2), p.String())
}

func Test_Protocol_Valid(t *testing.T) {
	var p Protocol

	p = HookProtocolDefault
	require.True(t, p.Valid())

	p = HookProtocolV2
	require.True(t, p.Valid())

	p = "invalid_protocol"
	require.False(t, p.Valid())
}
