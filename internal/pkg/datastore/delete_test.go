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

package datastore

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDatastoreDeleteArguments(t *testing.T) {
	mockAppID := "A0123456"

	for name, tt := range map[string]struct {
		Expression string
		Query      types.AppDatastoreDelete
		Results    types.AppDatastoreDeleteResult
	}{
		"Delete an item by ID": {
			Query: types.AppDatastoreDelete{
				Datastore: "Todos",
				App:       mockAppID,
				ID:        "1234",
			},
			Results: types.AppDatastoreDeleteResult{
				Datastore: "Todos",
				ID:        "1234",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			log := logger.Logger{
				Data: map[string]interface{}{},
			}
			clientsMock.APIInterface.On("AppsDatastoreDelete", mock.Anything, mock.Anything, tt.Query).
				Return(tt.Results, nil)
			client := shared.NewClientFactory(clientsMock.MockClientFactory())

			event, err := Delete(ctx, client, &log, tt.Query)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.Results, event.Data["deleteResult"])
			}
		})
	}
}
