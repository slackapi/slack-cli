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

package cache

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Manifest(t *testing.T) {
	tests := map[string]struct {
		mockAppID    string
		mockCache    ManifestCacheApp
		expectedHash Hash
	}{
		"missing cache entries return an empty hash": {
			mockAppID:    "A123",
			expectedHash: Hash(""),
		},
		"existing cache entries return the hash": {
			mockAppID:    "A123",
			mockCache:    ManifestCacheApp{Hash: Hash("xoxo")},
			expectedHash: Hash("xoxo"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			projectDirPath := "/path/to/project-name"
			err := fsMock.MkdirAll(filepath.Dir(projectDirPath), 0o755)
			require.NoError(t, err)
			cache := NewCache(fsMock, osMock, projectDirPath)
			ctx := context.Background()
			err = cache.SetManifestHash(ctx, tt.mockAppID, tt.mockCache.Hash)
			require.NoError(t, err)
			cache.ManifestCache.Apps = map[string]ManifestCacheApp{
				tt.mockAppID: tt.mockCache,
			}
			hash, err := cache.GetManifestHash(ctx, tt.mockAppID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHash, hash)
		})
	}
}

func TestCache_Manifest_NewManifestHash(t *testing.T) {
	tests := map[string]struct {
		mockManifest types.AppManifest
		expectedHash Hash
	}{
		"empty manifests hash to a constant": {
			mockManifest: types.AppManifest{},
			expectedHash: "d29ed745b10893910417522611d637aaabfe1c80aec059ac4c43efcb7e38f33c",
		},
		"custom manifest values hash to a different value": {
			mockManifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "slackbot[bot]",
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.SLACK_HOSTED,
					EventSubscriptions: &types.ManifestEventSubscriptions{
						BotEvents:  []string{"chat:write"},
						UserEvents: []string{"channels:read"},
					},
				},
			},
			expectedHash: "49691953b3bb36cad1333949846ad9f9c1fde9f12a395674dd2bbdafabccdd0c",
		},
		"reordered manifest values return the same hash": {
			mockManifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "slackbot[bot]",
				},
				Settings: &types.AppSettings{
					EventSubscriptions: &types.ManifestEventSubscriptions{
						UserEvents: []string{"channels:read"},
						BotEvents:  []string{"chat:write"},
					},
					FunctionRuntime: types.SLACK_HOSTED,
				},
			},
			expectedHash: "49691953b3bb36cad1333949846ad9f9c1fde9f12a395674dd2bbdafabccdd0c",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			projectDirPath := "/path/to/project-name"
			err := fsMock.MkdirAll(filepath.Dir(projectDirPath), 0o755)
			require.NoError(t, err)
			ctx := context.Background()
			cache := NewCache(fsMock, osMock, projectDirPath)
			hash, err := cache.NewManifestHash(ctx, tt.mockManifest)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHash, hash)
		})
	}
}
