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
	"slices"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// LoadExperiments parses experiments from the command flags and configuration
// files and stores the findings in Config
func (c *Config) LoadExperiments(
	ctx context.Context,
	printDebug func(ctx context.Context, format string, a ...interface{}),
) {
	experiments := []experiment.Experiment{}
	// Load from flags
	for _, flagValue := range c.ExperimentsFlag {
		experiments = append(experiments, experiment.Experiment(flagValue))
	}
	printDebug(ctx, fmt.Sprintf("active flag experiments: %s", experiments))
	// Load from project config file
	projectConfig, err := ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrInvalidAppDirectory {
		printDebug(ctx, fmt.Sprintf("failed to parse project-level config file: %s", err))
	} else if err == nil {
		printDebug(ctx, fmt.Sprintf("active project experiments: %s", projectConfig.Experiments))
		experiments = append(experiments, projectConfig.Experiments...)
	}
	// Load from system config file
	userConfig, err := c.SystemConfig.UserConfig(ctx)
	if err != nil {
		printDebug(ctx, fmt.Sprintf("failed to parse system-level config file: %s", err))
	} else {
		printDebug(ctx, fmt.Sprintf("active system experiments: %s", userConfig.Experiments))
		experiments = append(experiments, userConfig.Experiments...)
	}
	// Load from permanently enabled list
	experiments = append(experiments, experiment.EnabledExperiments...)
	printDebug(ctx, fmt.Sprintf("active permanently enabled experiments: %s", experiment.EnabledExperiments))
	// Sort, trim, and audit the experiments list
	slices.Sort(experiments)
	c.experiments = slices.Compact(experiments)
	for _, exp := range c.experiments {
		if !experiment.Includes(exp) {
			printDebug(ctx, fmt.Sprintf("invalid experiment found: %s", exp))
		}
	}
}

// GetExperiments returns the set of active experiments
func (c *Config) GetExperiments() []experiment.Experiment {
	return c.experiments
}

// WithExperimentOn checks whether an experiment is currently toggled on
func (c *Config) WithExperimentOn(experimentToCheck experiment.Experiment) bool {
	return slices.Contains(c.experiments, experimentToCheck)
}
