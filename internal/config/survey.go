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

	"github.com/opentracing/opentracing-go"
)

// SurveyConfig tracks information related to user surveys in the system-level or project-level config
type SurveyConfig struct {
	AskedAt     string `json:"asked_at"`
	CompletedAt string `json:"completed_at"`
}

// GetSurveyConfig returns the survey for the given survey ID.
// It combines survey config at the project-level and system-level.
// If the survey ID does not exist, an error is returned.
func (c *Config) GetSurveyConfig(ctx context.Context, name string) (SurveyConfig, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetSurveyConfig")
	defer span.Finish()

	// First check project config
	survey, err := c.ProjectConfig.GetSurveyConfig(ctx, name)
	if err == nil { // If fetched successfully, return now
		return survey, nil
	}

	// Not able to fetch from project config; try system config
	survey, err = c.SystemConfig.GetSurveyConfig(ctx, name)
	if err != nil {
		return SurveyConfig{}, err
	}

	return survey, nil
}
