package manifest

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Flatten converts a manifest (as JSON-serializable struct) into a flat map
// where keys are dot-notation paths and values are the leaf values.
// Arrays are treated as leaf values (not recursed into individually) because
// array element identity is ambiguous without a key field.
func Flatten(manifest any) (map[string]any, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	result := make(map[string]any)
	flattenRecursive("", raw, result)
	return result, nil
}

func flattenRecursive(prefix string, data map[string]any, result map[string]any) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]any:
			flattenRecursive(fullKey, v, result)
		default:
			result[fullKey] = value
		}
	}
}

// SortedKeys returns the keys of a flat map in sorted order for deterministic output.
func SortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
