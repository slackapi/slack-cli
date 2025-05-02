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

func TestDatastoreQueryArguments(t *testing.T) {
	mockAppID := "A0123456"

	for name, tt := range map[string]struct {
		Expression string
		Query      types.AppDatastoreQuery
		Results    types.AppDatastoreQueryResult
	}{
		"Empty expression without limits": {
			Query: types.AppDatastoreQuery{
				Datastore: "Todos",
				App:       mockAppID,
			},
			Results: types.AppDatastoreQueryResult{
				Datastore: "Todos",
				Items: []map[string]interface{}{
					{"id": "1", "name": "drink water", "status": "done"},
					{"id": "2", "name": "write tests", "status": "done"},
					{"id": "3", "name": "take a walk", "status": "soon"},
				},
			},
		},
		"Empty expression with limits": {
			Query: types.AppDatastoreQuery{
				Datastore: "Todos",
				App:       mockAppID,
				Limit:     1,
			},
			Results: types.AppDatastoreQueryResult{
				Datastore: "Todos",
				Items: []map[string]interface{}{
					{"id": "1", "name": "drink water", "status": "done"},
				},
				NextCursor: "arandomcursorhere",
			},
		},
		"Expression with a limit that isn't met": {
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
				Expression: "#status = :status",
				ExpressionAttributes: map[string]interface{}{
					"#status": "status",
				},
				ExpressionValues: map[string]interface{}{
					":status": "soon",
				},
				Limit: 2,
			},
			Results: types.AppDatastoreQueryResult{
				Datastore: "Todos",
				Items: []map[string]interface{}{
					{"id": "3", "name": "take a walk", "status": "soon"},
				},
			},
		},
		"Expression with a limit and cursor": {
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
				Expression: "#status = :status",
				ExpressionAttributes: map[string]interface{}{
					"#status": "status",
				},
				ExpressionValues: map[string]interface{}{
					":status": "done",
				},
				Limit:  1,
				Cursor: "arandomcursorhere",
			},
			Results: types.AppDatastoreQueryResult{
				Datastore: "Todos",
				Items: []map[string]interface{}{
					{"id": "2", "name": "write tests", "status": "done"},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			log := logger.Logger{
				Data: map[string]interface{}{},
			}
			clientsMock.APIInterface.On("AppsDatastoreQuery", mock.Anything, mock.Anything, tt.Query).
				Return(tt.Results, nil)
			client := shared.NewClientFactory(clientsMock.MockClientFactory())

			event, err := Query(ctx, client, &log, tt.Query)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.Results, event.Data["queryResult"])
			}
		})
	}
}
