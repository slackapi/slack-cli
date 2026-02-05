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
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
)

func Test_Config_GetSurveyConfig(t *testing.T) {
	tests := map[string]struct {
		surveyName     string
		expectedSurvey SurveyConfig
		expectedError  *slackerror.Error
	}{
		"errors if the survey configurations cannot be found": {
			surveyName:    "air",
			expectedError: slackerror.New(slackerror.ErrSurveyConfigNotFound),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, _, _, c, _, _ := setup(t)
			surveyConfig, err := c.GetSurveyConfig(ctx, tc.surveyName)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.Equal(t, tc.expectedSurvey, surveyConfig)
			}
		})
	}
}
