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

package config

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config_WithExperimentOn(t *testing.T) {
	validExperiment := experiment.Placeholder
	invalidExperiment := "everythingShouldFail"

	t.Run("Correctly finds experiments when they are in config.json", func(t *testing.T) {
		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()

		// Write our test script to the memory file system used by the mock
		jsonContents := []byte(fmt.Sprintf(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{"%s":true}}`, validExperiment))
		_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)

		config.LoadExperiments(ctx, mockPrintDebug)
		experimentOn := config.WithExperimentOn(validExperiment)
		assert.Equal(t, true, experimentOn)
		assert.Contains(t, mockOutput.String(),
			fmt.Sprintf("active system experiments: [%s]", validExperiment))
	})

	t.Run("Returns a parse error with remediation for old array format in config.json", func(t *testing.T) {
		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()

		// Write old array format
		jsonContents := []byte(
			fmt.Sprintf(
				`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":["%s"]}`,
				validExperiment,
			),
		)
		_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)

		config.LoadExperiments(ctx, mockPrintDebug)
		experimentOn := config.WithExperimentOn(validExperiment)
		assert.Equal(t, false, experimentOn)
		assert.Contains(t, mockOutput.String(), "failed to parse system-level config file")
	})

	t.Run("Correctly returns false when experiments are not in config.json", func(t *testing.T) {
		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()

		jsonContents := []byte(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{}}`)
		_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)

		config.LoadExperiments(ctx, mockPrintDebug)
		experimentOn := config.WithExperimentOn(validExperiment)
		assert.Equal(t, false, experimentOn)
		assert.NotContains(t, mockOutput.String(), validExperiment)
	})

	t.Run("Correctly finds experiments when set via experiment flag", func(t *testing.T) {
		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()

		// Write no contents via config.json
		jsonContents := []byte(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{}}`)
		_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)
		config.ExperimentsFlag = []string{string(validExperiment)}

		config.LoadExperiments(ctx, mockPrintDebug)
		experimentOn := config.WithExperimentOn(validExperiment)
		assert.Equal(t, true, experimentOn)
		assert.Contains(t, mockOutput.String(),
			fmt.Sprintf("active flag experiments: [%s]", validExperiment))
	})

	t.Run("Correctly logs error when experiments set via config are invalid", func(t *testing.T) {
		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()

		jsonContents := []byte(fmt.Sprintf(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{"%s":true}}`, invalidExperiment))
		_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)
		config.LoadExperiments(ctx, mockPrintDebug)
		assert.Contains(t, mockOutput.String(),
			fmt.Sprintf("invalid experiment found: %s", invalidExperiment))
	})

	t.Run("Correctly validates valid experiments", func(t *testing.T) {
		tableTests := map[string]struct {
			experiment  string
			expectedRes bool
		}{
			"valid experiments are valid": {
				experiment:  string(validExperiment),
				expectedRes: true,
			},
			"invalid experiments are invalid": {
				experiment:  invalidExperiment,
				expectedRes: false,
			},
		}

		// set via config.json
		for name, tc := range tableTests {
			t.Run(name, func(t *testing.T) {
				// Setup
				ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
				defer teardown(t)
				mockOutput, mockPrintDebug := setupMockPrintDebug()

				// Write contents via config.json
				jsonContents := []byte(fmt.Sprintf(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{"%s":true}}`, tc.experiment))
				_ = afero.WriteFile(fs, pathToConfigJSON, jsonContents, 0600)

				config.LoadExperiments(ctx, mockPrintDebug)
				if !tc.expectedRes {
					assert.Contains(t, mockOutput.String(),
						fmt.Sprintf("invalid experiment found: %s", tc.experiment))
				} else {
					isActive := config.WithExperimentOn(experiment.Experiment(tc.experiment))
					assert.Equal(t, tc.expectedRes, isActive)
					assert.Contains(t, mockOutput.String(),
						fmt.Sprintf("active system experiments: [%s]", tc.experiment))
				}
			})
		}

		// set via flag
		for name, tc := range tableTests {
			t.Run(name, func(t *testing.T) {
				// Setup
				ctx, _, _, config, _, teardown := setup(t)
				defer teardown(t)
				mockOutput, mockPrintDebug := setupMockPrintDebug()

				// Add environment variables via experiment flag
				config.ExperimentsFlag = []string{tc.experiment}

				// look for a match which is invalid format
				config.LoadExperiments(ctx, mockPrintDebug)
				if !tc.expectedRes {
					assert.Contains(t, mockOutput.String(),
						fmt.Sprintf("invalid experiment found: %s", tc.experiment))
				} else {
					isActive := config.WithExperimentOn(experiment.Experiment(tc.experiment))
					assert.Equal(t, tc.expectedRes, isActive)
					assert.Contains(t, mockOutput.String(),
						fmt.Sprintf("active flag experiments: [%s]", tc.experiment))
				}
			})
		}
	})

	t.Run("Loads valid experiments from project configs", func(t *testing.T) {
		ctx, fs, _, config, _, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()
		err := fs.Mkdir(slackdeps.MockWorkingDirectory, 0755)
		require.NoError(t, err)
		err = fs.Mkdir(filepath.Join(slackdeps.MockWorkingDirectory, ProjectConfigDirName), 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0600)
		require.NoError(t, err)
		jsonContents := []byte(fmt.Sprintf(`{"experiments":{"%s":true}}`, experiment.Placeholder))
		err = afero.WriteFile(fs, GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory), []byte(jsonContents), 0600)
		require.NoError(t, err)

		config.LoadExperiments(ctx, mockPrintDebug)
		assert.True(t, config.WithExperimentOn(experiment.Placeholder))
		assert.Contains(t, mockOutput.String(),
			fmt.Sprintf("active project experiments: [%s]", experiment.Placeholder))
	})

	t.Run("Loads valid experiments from project configs and removes duplicates", func(t *testing.T) {
		ctx, fs, _, config, _, teardown := setup(t)
		defer teardown(t)
		_, mockPrintDebug := setupMockPrintDebug()
		err := fs.Mkdir(slackdeps.MockWorkingDirectory, 0755)
		require.NoError(t, err)
		err = fs.Mkdir(filepath.Join(slackdeps.MockWorkingDirectory, ProjectConfigDirName), 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0600)
		require.NoError(t, err)
		jsonContents := []byte(
			fmt.Sprintf(`{"experiments":{"%s":true, "%s":true, "%s":true}}`,
				experiment.Placeholder,
				experiment.Placeholder,
				experiment.Placeholder,
			),
		)
		err = afero.WriteFile(fs, GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory), []byte(jsonContents), 0600)
		require.NoError(t, err)

		// Also set via flag to test dedup
		config.ExperimentsFlag = []string{string(experiment.Placeholder)}

		config.LoadExperiments(ctx, mockPrintDebug)
		assert.True(t, config.WithExperimentOn(experiment.Placeholder))
		// Assert that experiments are deduplicated via map
		exps := config.GetExperiments()
		count := 0
		for _, exp := range exps {
			if exp == experiment.Placeholder {
				count++
			}
		}
		assert.Equal(t, 1, count, "experiment should appear exactly once")
	})

	t.Run("Loads valid experiments and enabled default experiments", func(t *testing.T) {
		ctx, fs, _, config, _, teardown := setup(t)
		defer teardown(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()
		err := fs.Mkdir(slackdeps.MockWorkingDirectory, 0755)
		require.NoError(t, err)
		err = fs.Mkdir(filepath.Join(slackdeps.MockWorkingDirectory, ProjectConfigDirName), 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0600)
		require.NoError(t, err)
		jsonContents := []byte(`{"experiments":{}}`) // No experiments
		err = afero.WriteFile(fs, GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory), []byte(jsonContents), 0600)
		require.NoError(t, err)

		// Add a default experiment, because there is no guarantee to be one
		var _EnabledExperiments = experiment.EnabledExperiments
		defer func() {
			// Restore original EnabledExperiments
			experiment.EnabledExperiments = _EnabledExperiments
		}()
		experiment.EnabledExperiments = []experiment.Experiment{experiment.Placeholder} // Placeholder enabled by default

		// Run test
		config.LoadExperiments(ctx, mockPrintDebug)
		assert.True(t, config.WithExperimentOn(experiment.Placeholder))
		assert.Contains(t, mockOutput.String(), "active project experiments: []")
		assert.Contains(t, mockOutput.String(), fmt.Sprintf("active permanently enabled experiments: [%s]", experiment.Placeholder))
		assert.ElementsMatch(t, []experiment.Experiment{experiment.Placeholder}, config.GetExperiments())
	})

	t.Run("Logs an invalid experiments in project configs", func(t *testing.T) {
		ctx, fs, _, config, _, teardown := setup(t)
		mockOutput, mockPrintDebug := setupMockPrintDebug()
		defer teardown(t)
		err := fs.Mkdir(slackdeps.MockWorkingDirectory, 0755)
		require.NoError(t, err)
		err = fs.Mkdir(filepath.Join(slackdeps.MockWorkingDirectory, ProjectConfigDirName), 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0600)
		require.NoError(t, err)
		jsonContents := []byte(`{"experiments":{"experiment-37":true}}`)
		err = afero.WriteFile(fs, GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory), []byte(jsonContents), 0600)
		require.NoError(t, err)

		config.LoadExperiments(ctx, mockPrintDebug)
		assert.Contains(t, mockOutput.String(), "invalid experiment found: experiment-37")
	})
}

func Test_Config_GetExperiments(t *testing.T) {
	validExperiment := experiment.Placeholder
	t.Run("Should get unique valid experiments only", func(t *testing.T) {
		// Remove any enabled experiments during the test and restore afterward
		var _EnabledExperiments = experiment.EnabledExperiments
		experiment.EnabledExperiments = []experiment.Experiment{}
		defer func() {
			// Restore original EnabledExperiments
			experiment.EnabledExperiments = _EnabledExperiments
		}()

		// Setup
		ctx, fs, _, config, pathToConfigJSON, teardown := setup(t)
		defer teardown(t)
		_, mockPrintDebug := setupMockPrintDebug()

		// Write contents via config.json
		var configJSON = []byte(fmt.Sprintf(`{"last_update_checked_at":"2023-05-11T15:41:07.799619-07:00","experiments":{"%s":true}}`, validExperiment))
		_ = afero.WriteFile(fs, pathToConfigJSON, configJSON, 0600)

		// Set contexts of experiment flag
		// Add environment variables via experiment flag
		config.ExperimentsFlag = []string{string(validExperiment), string(validExperiment)}

		config.LoadExperiments(ctx, mockPrintDebug)
		exp := config.GetExperiments()
		assert.ElementsMatch(t, []experiment.Experiment{validExperiment}, exp)
	})
}

// setupMockPrintDebug prepares a stubbed writer to avoid circular imports
func setupMockPrintDebug() (*bytes.Buffer, func(context.Context, string, ...interface{})) {
	mockOutput := &bytes.Buffer{}
	mockPrintDebug := func(ctx context.Context, format string, a ...interface{}) {
		fmt.Fprintf(mockOutput, format, a...)
	}
	return mockOutput, mockPrintDebug
}
