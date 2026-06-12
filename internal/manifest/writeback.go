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

package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
)

const manifestFileName = "manifest.json"

// WriteBackResult describes what happened during write-back.
type WriteBackResult struct {
	Written  bool
	FilePath string
	Warning  string
}

// WriteManifestLocal writes the merged manifest back to the project's
// manifest.json file, preserving the original file's key ordering by
// using the same JSON structure.
func WriteManifestLocal(fs afero.Fs, workingDir string, manifest types.AppManifest) (WriteBackResult, error) {
	manifestPath := filepath.Join(workingDir, manifestFileName)

	exists, err := afero.Exists(fs, manifestPath)
	if err != nil {
		return WriteBackResult{}, fmt.Errorf("failed to check manifest file: %w", err)
	}
	if !exists {
		return WriteBackResult{
			Warning: fmt.Sprintf("No %s found in project root — merged manifest was not written locally", manifestFileName),
		}, nil
	}

	original, err := afero.ReadFile(fs, manifestPath)
	if err != nil {
		return WriteBackResult{}, fmt.Errorf("failed to read %s: %w", manifestFileName, err)
	}

	merged, fellBack, err := marshalPreservingOrder(original, manifest)
	if err != nil {
		return WriteBackResult{}, fmt.Errorf("failed to serialize merged manifest: %w", err)
	}

	if err := atomicWriteFile(fs, manifestPath, merged, 0644); err != nil {
		return WriteBackResult{}, fmt.Errorf("failed to write %s: %w", manifestFileName, err)
	}

	result := WriteBackResult{Written: true, FilePath: manifestPath}
	if fellBack {
		result.Warning = fmt.Sprintf("Could not parse the original %s, so its key order was not preserved", manifestFileName)
	}
	return result, nil
}

// atomicWriteFile writes to a sibling temp file and renames it over the
// destination so an interrupted write cannot leave the destination truncated.
func atomicWriteFile(fs afero.Fs, dest string, data []byte, mode os.FileMode) error {
	tmp := dest + ".tmp"
	if err := afero.WriteFile(fs, tmp, data, mode); err != nil {
		return err
	}
	if err := fs.Rename(tmp, dest); err != nil {
		_ = fs.Remove(tmp)
		return err
	}
	return nil
}

// marshalPreservingOrder serializes the manifest to JSON, preserving the
// top-level key order from the original file. The returned fellBack is true
// when the original could not be parsed and the result was emitted with
// default key order instead.
func marshalPreservingOrder(original []byte, manifest types.AppManifest) (data []byte, fellBack bool, err error) {
	var originalKeys []string
	if err := extractTopLevelKeyOrder(original, &originalKeys); err != nil {
		fresh, err := marshalFresh(manifest)
		return fresh, true, err
	}

	newData, err := json.Marshal(manifest)
	if err != nil {
		return nil, false, err
	}
	var newMap map[string]json.RawMessage
	if err := json.Unmarshal(newData, &newMap); err != nil {
		fresh, err := marshalFresh(manifest)
		return fresh, true, err
	}

	// Build output with original key order, then append any new keys
	type kv struct {
		Key   string
		Value json.RawMessage
	}
	var ordered []kv
	seen := make(map[string]bool)

	for _, key := range originalKeys {
		if val, exists := newMap[key]; exists {
			ordered = append(ordered, kv{Key: key, Value: val})
			seen[key] = true
		}
	}
	for key, val := range newMap {
		if !seen[key] {
			ordered = append(ordered, kv{Key: key, Value: val})
		}
	}

	// Manually build JSON with preserved order
	buf := []byte("{\n")
	for i, item := range ordered {
		keyJSON, err := json.Marshal(item.Key)
		if err != nil {
			return nil, false, fmt.Errorf("failed to marshal manifest key %q: %w", item.Key, err)
		}
		indented, err := json.MarshalIndent(json.RawMessage(item.Value), "  ", "  ")
		if err != nil {
			return nil, false, fmt.Errorf("failed to marshal manifest value at %q: %w", item.Key, err)
		}
		buf = fmt.Appendf(buf, "  %s: %s", keyJSON, indented)
		if i < len(ordered)-1 {
			buf = append(buf, ',')
		}
		buf = append(buf, '\n')
	}
	buf = append(buf, '}')
	buf = append(buf, '\n')

	return buf, false, nil
}

func extractTopLevelKeyOrder(data []byte, keys *[]string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// json.Unmarshal into map doesn't preserve order, so we need to parse tokens
	dec := json.NewDecoder(bytes.NewReader(data))
	// Read opening brace
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected opening brace")
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := t.(string)
		if !ok {
			continue
		}
		*keys = append(*keys, key)
		// Skip the value
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return err
		}
	}
	return nil
}

func marshalFresh(manifest types.AppManifest) ([]byte, error) {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
