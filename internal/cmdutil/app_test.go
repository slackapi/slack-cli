// Copyright 2022-2026 Salesforce, Inc.
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

package cmdutil

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/assert"
)

func TestAppExists(t *testing.T) {
	localAppExists := types.App{
		AppID:      "A12345",
		IsDev:      true,
		TeamDomain: "test",
		TeamID:     "T12345",
	}

	localAppDoesNotExist := types.App{
		IsDev:      true,
		TeamDomain: "test",
		TeamID:     "T12345",
	}

	hostedAppExists := types.App{
		AppID:      "A12345",
		TeamDomain: "test",
		TeamID:     "T12345",
	}

	hostedAppDoesNotExist := types.App{
		TeamDomain: "test",
		TeamID:     "T12345",
	}

	mockAuth := types.SlackAuth{
		TeamDomain: "test",
		TeamID:     "T12345",
		Token:      "xoxp",
	}

	res1 := AppExists(localAppExists, mockAuth)
	res2 := AppExists(hostedAppExists, mockAuth)
	res3 := AppExists(localAppDoesNotExist, mockAuth)
	res4 := AppExists(hostedAppDoesNotExist, mockAuth)
	assert.NoError(t, res1)
	assert.NoError(t, res2)
	assert.Error(t, res3)
	assert.Error(t, res4)
}
