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

func TestDatastoreGetArguments(t *testing.T) {
	mockAppID := "A0123456"

	for name, tc := range map[string]struct {
		Expression string
		Query      types.AppDatastoreGet
		Results    types.AppDatastoreGetResult
	}{
		"Get an item by ID": {
			Query: types.AppDatastoreGet{
				Datastore: "Todos",
				App:       mockAppID,
				ID:        "2",
			},
			Results: types.AppDatastoreGetResult{
				Datastore: "Todos",
				Item:      map[string]interface{}{"id": "2", "name": "write tests", "status": "done"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			log := logger.Logger{
				Data: map[string]interface{}{},
			}
			clientsMock.API.On("AppsDatastoreGet", mock.Anything, mock.Anything, tc.Query).
				Return(tc.Results, nil)
			client := shared.NewClientFactory(clientsMock.MockClientFactory())

			event, err := Get(ctx, client, &log, tc.Query)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.Results, event.Data["getResult"])
			}
		})
	}
}
