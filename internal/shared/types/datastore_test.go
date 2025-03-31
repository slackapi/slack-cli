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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Datastore(t *testing.T) {
	tests := map[string]struct {
		query   Datastorer
		name    string
		setName string
		appID   string
	}{
		"query expressions resolve": {
			query: &AppDatastoreQuery{Datastore: "D0", App: "A0"},
			name:  "D0",
			appID: "A0",
		},
		"query name can be set": {
			query:   &AppDatastoreQuery{},
			name:    "D0",
			setName: "D0",
		},
		"count expressions resolve": {
			query: &AppDatastoreCount{Datastore: "D1", App: "A1"},
			name:  "D1",
			appID: "A1",
		},
		"count name can be set": {
			query:   &AppDatastoreCount{},
			name:    "D1",
			setName: "D1",
		},
		"put expressions resolve": {
			query: &AppDatastorePut{Datastore: "D2", App: "A2"},
			name:  "D2",
			appID: "A2",
		},
		"put name can be set": {
			query:   &AppDatastorePut{},
			name:    "D2",
			setName: "D2",
		},
		"bulk put expressions resolve": {
			query: &AppDatastoreBulkPut{Datastore: "D3", App: "A3"},
			name:  "D3",
			appID: "A3",
		},
		"bulk put name can be set": {
			query:   &AppDatastoreBulkPut{},
			name:    "D3",
			setName: "D3",
		},
		"update expressions resolve": {
			query: &AppDatastoreUpdate{Datastore: "D4", App: "A4"},
			name:  "D4",
			appID: "A4",
		},
		"update name can be set": {
			query:   &AppDatastoreUpdate{},
			name:    "D4",
			setName: "D4",
		},
		"delete expressions resolve": {
			query: &AppDatastoreDelete{Datastore: "D5", App: "A5"},
			name:  "D5",
			appID: "A5",
		},
		"delete name can be set": {
			query:   &AppDatastoreDelete{},
			name:    "D5",
			setName: "D5",
		},
		"bulk delete expressions resolve": {
			query: &AppDatastoreBulkDelete{Datastore: "D6", App: "A6"},
			name:  "D6",
			appID: "A6",
		},
		"bulk delete name can be set": {
			query:   &AppDatastoreBulkDelete{},
			name:    "D6",
			setName: "D6",
		},
		"get expressions resolve": {
			query: &AppDatastoreGet{Datastore: "D7", App: "A7"},
			name:  "D7",
			appID: "A7",
		},
		"get name can be set": {
			query:   &AppDatastoreGet{},
			name:    "D7",
			setName: "D7",
		},
		"bulk get expressions resolve": {
			query: &AppDatastoreBulkGet{Datastore: "D8", App: "A8"},
			name:  "D8",
			appID: "A8",
		},
		"bulk get name can be set": {
			query:   &AppDatastoreBulkGet{},
			name:    "D8",
			setName: "D8",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.setName != "" {
				tt.query.SetName(tt.setName)
			}
			assert.Equal(t, tt.name, tt.query.Name())
			assert.Equal(t, tt.appID, tt.query.AppID())
		})
	}
}
