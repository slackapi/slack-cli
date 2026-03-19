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
	"context"
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// experimentsFormatHint is a remediation hint for config files that use the
// deprecated array format for experiments instead of the current object format.
const experimentsFormatHint = `If the "experiments" field is an array (e.g. ["exp1"]), update it to an object (e.g. {"exp1": true})`

// LoadExperiments parses experiments from the command flags and configuration
// files and stores the findings in Config
func (c *Config) LoadExperiments(
	ctx context.Context,
	printDebug func(ctx context.Context, format string, a ...interface{}),
) {
	experiments := map[experiment.Experiment]bool{}
	// Load from permanently enabled list (lowest priority)
	for _, exp := range experiment.EnabledExperiments {
		experiments[exp] = true
	}
	printDebug(ctx, fmt.Sprintf("active permanently enabled experiments: %s", experiment.EnabledExperiments))
	// Load from system config file
	userConfig, err := c.SystemConfig.UserConfig(ctx)
	if err != nil {
		printDebug(ctx, fmt.Sprintf("failed to parse system-level config file: %s", err))
	} else {
		printDebug(ctx, fmt.Sprintf("active system experiments: %s", formatExperimentMap(toExperimentMap(userConfig.Experiments))))
		maps.Copy(experiments, toExperimentMap(userConfig.Experiments))
	}
	// Load from project config file
	projectConfig, err := ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrInvalidAppDirectory {
		printDebug(ctx, fmt.Sprintf("failed to parse project-level config file: %s", err))
	} else if err == nil {
		printDebug(ctx, fmt.Sprintf("active project experiments: %s", formatExperimentMap(toExperimentMap(projectConfig.Experiments))))
		maps.Copy(experiments, toExperimentMap(projectConfig.Experiments))
	}
	// Load from flags (highest priority)
	for _, flagValue := range c.ExperimentsFlag {
		experiments[experiment.Experiment(flagValue)] = true
	}
	printDebug(ctx, fmt.Sprintf("active flag experiments: %s", formatExperimentMap(experiments)))
	// Audit the experiments
	c.experiments = experiments
	for name := range c.experiments {
		if !experiment.Includes(name) {
			printDebug(ctx, fmt.Sprintf("invalid experiment found: %s", name))
		}
	}
}

// GetExperiments returns the set of active experiments
func (c *Config) GetExperiments() []experiment.Experiment {
	result := make([]experiment.Experiment, 0, len(c.experiments))
	for name, enabled := range c.experiments {
		if enabled {
			result = append(result, name)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

// WithExperimentOn checks whether an experiment is currently toggled on
func (c *Config) WithExperimentOn(experimentToCheck experiment.Experiment) bool {
	return c.experiments[experimentToCheck]
}

// toExperimentMap converts a map[string]bool to map[experiment.Experiment]bool
func toExperimentMap(m map[string]bool) map[experiment.Experiment]bool {
	result := make(map[experiment.Experiment]bool, len(m))
	for name, enabled := range m {
		result[experiment.Experiment(name)] = enabled
	}
	return result
}

// formatExperimentMap returns a string representation of the experiments map
// for debug logging, formatted similar to the old slice format.
func formatExperimentMap(m map[experiment.Experiment]bool) string {
	if len(m) == 0 {
		return "[]"
	}
	names := make([]string, 0, len(m))
	for name, enabled := range m {
		if enabled {
			names = append(names, string(name))
		}
	}
	sort.Strings(names)
	return "[" + strings.Join(names, " ") + "]"
}
