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

package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setup
func setup(t *testing.T) (context.Context, *slackdeps.FsMock, *slackdeps.OsMock, *Config, string, func(*testing.T)) {
	ctx := slackcontext.MockContext(t.Context())
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	os.AddDefaultMocks()
	config := NewConfig(fs, os)
	dir, _ := os.UserHomeDir()
	pathToConfigJSON := filepath.Join(dir, ".slack/config.json")

	return ctx, fs, os, config, pathToConfigJSON, func(t *testing.T) {
		_ = fs.Remove(pathToConfigJSON)
	}
}

func Test_Config_NewConfig(t *testing.T) {
	t.Run("NewConfig(fs, os) when arguments nil", func(t *testing.T) {
		var fs afero.Fs
		var os types.Os
		fs = nil
		os = nil

		var config = NewConfig(fs, os)
		require.Equal(t, nil, config.fs)
		require.Equal(t, nil, config.os)
	})

	// Test: when arguments exist
	t.Run("NewConfig(fs, os) when arguments exist", func(t *testing.T) {
		var fs = slackdeps.NewFsMock()
		var os = slackdeps.NewOsMock()
		os.AddDefaultMocks()

		var config = NewConfig(fs, os)
		require.Equal(t, fs, config.fs)
		require.Equal(t, os, config.os)
	})
}

func Test_Config_SkipLocalFs(t *testing.T) {
	tests := map[string]struct {
		tokenFlag  string
		appFlag    string
		isHeadless bool
	}{
		"interactive when no flag values are set": {
			tokenFlag:  "",
			appFlag:    "",
			isHeadless: false,
		},
		"interactive with only a token flag": {
			tokenFlag:  "xoxp-example-token",
			appFlag:    "",
			isHeadless: false,
		},
		"interactive with only an app flag": {
			tokenFlag:  "",
			appFlag:    "A0123456789",
			isHeadless: false,
		},
		"interactive with a token and app environment flag": {
			tokenFlag:  "xoxp-example-token",
			appFlag:    "local",
			isHeadless: false,
		},
		"headless with an app id and token flag": {
			tokenFlag:  "xoxp-example-token",
			appFlag:    "A0123456789",
			isHeadless: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			config := NewConfig(fs, os)
			config.TokenFlag = tc.tokenFlag
			config.AppFlag = tc.appFlag

			isHeadless := config.SkipLocalFs()
			assert.Equal(t, tc.isHeadless, isHeadless)
		})
	}
}
