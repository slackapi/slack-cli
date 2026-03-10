package prompts

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessLabels(t *testing.T) {
	tests := map[string]struct {
		current         types.Permission
		expectedCurrent string
	}{
		"app_collaborators as current": {
			current:         types.PermissionAppCollaborators,
			expectedCurrent: "app collaborators only (current)",
		},
		"everyone as current": {
			current:         types.PermissionEveryone,
			expectedCurrent: "everyone (current)",
		},
		"named_entities as current": {
			current:         types.PermissionNamedEntities,
			expectedCurrent: "specific users (current)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			labels, distributions := AccessLabels(tc.current)
			assert.Len(t, labels, 3)
			assert.Len(t, distributions, 3)
			assert.Equal(t, tc.expectedCurrent, labels[0])
			assert.Equal(t, tc.current, distributions[0])
		})
	}
}

func TestTriggerAccessLabels(t *testing.T) {
	tests := map[string]struct {
		current         types.Permission
		expectedCurrent string
	}{
		"app_collaborators as current": {
			current:         types.PermissionAppCollaborators,
			expectedCurrent: "app collaborators only (current)",
		},
		"everyone as current": {
			current:         types.PermissionEveryone,
			expectedCurrent: "everyone (current)",
		},
		"named_entities as current": {
			current:         types.PermissionNamedEntities,
			expectedCurrent: "specific entities (current)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			labels, distributions := TriggerAccessLabels(tc.current)
			assert.Len(t, labels, 3)
			assert.Len(t, distributions, 3)
			assert.Equal(t, tc.expectedCurrent, labels[0])
			assert.Equal(t, tc.current, distributions[0])
		})
	}
}
