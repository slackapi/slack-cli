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
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
)

// ManifestCacher saves and retrieves specific manifest values
type ManifestCacher interface {
	GetManifestHash(ctx context.Context, appID string) (Hash, error)
	NewManifestHash(ctx context.Context, manifest types.AppManifest) (Hash, error)
	SetManifestHash(ctx context.Context, appID string, hash Hash) error
}

// ManifestCache stores values of an app manifest
type ManifestCache struct {
	Apps map[string]ManifestCacheApp
}

// ManifestCacheApp contains cache details for a specific app manifest
type ManifestCacheApp struct {
	Hash Hash `json:"hash"` // Hash is a computed value unique to a manifest
}

// GetManifestHash loads the saved manifest hash from cache
func (c *Cache) GetManifestHash(ctx context.Context, appID string) (Hash, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetManifestHash")
	defer span.Finish()
	cache, err := c.readManifestCache(ctx)
	if err != nil {
		return "", err
	}
	return cache[appID].Hash, nil
}

// NewManifestHash creates a hash unique to the manifest
//
// The source of the manifest provided should be noted since values from hooks
// might not be the same as the API response.
func (c *Cache) NewManifestHash(ctx context.Context, manifest types.AppManifest) (Hash, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "NewManifestHash")
	defer span.Finish()
	bytes, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return NewHash(bytes), nil
}

// SetManifestHash saves the manifest hash for an app ID
func (c *Cache) SetManifestHash(ctx context.Context, appID string, hash Hash) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "SetManifestHash")
	defer span.Finish()
	cache, err := c.readManifestCache(ctx)
	if err != nil {
		return err
	}
	cache[appID] = ManifestCacheApp{
		Hash: hash,
	}
	c.ManifestCache.Apps = cache
	return c.writeManifestCache(ctx)
}

// readManifestCache loads the manifest cache from file
func (c *Cache) readManifestCache(ctx context.Context) (cache map[string]ManifestCacheApp, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "readManifestCache")
	defer span.Finish()

	// Prefer "manifests.json" over "manifest.json" to avoid updating the manifest
	// when changing files included in the "watch.filter-regex" hook configuration.
	path := filepath.Join(c.path, ".slack", "cache", "manifests.json")
	bytes, err := afero.ReadFile(c.fs, path)
	switch {
	case os.IsNotExist(err):
		return map[string]ManifestCacheApp{}, nil
	case err != nil:
		return map[string]ManifestCacheApp{}, err
	}
	err = json.Unmarshal(bytes, &cache)
	if err != nil {
		return map[string]ManifestCacheApp{}, err
	}
	return cache, nil
}

// writeManifestCache saves the manifest cache to file
func (c *Cache) writeManifestCache(ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "writeManifestCache")
	defer span.Finish()
	err := c.createCacheDir()
	if err != nil && !os.IsExist(err) {
		return err
	}
	cache, err := json.MarshalIndent(c.ManifestCache.Apps, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(c.path, ".slack", "cache", "manifests.json")
	err = afero.WriteFile(c.fs, path, cache, 0o644)
	if err != nil {
		return err
	}
	return nil
}
