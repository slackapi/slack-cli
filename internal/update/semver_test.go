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

package update

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
)

func Test_Update_SemVerGreaterThan(t *testing.T) {
	tests := map[string]struct {
		v1       string
		v2       string
		expected bool
		err      error
	}{
		"equal versions are not greater": {
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: false,
		},
		"greater than major is greater": {
			v1:       "2.0.0",
			v2:       "1.2.3",
			expected: true,
		},
		"greater than minor is greater": {
			v1:       "1.3.0",
			v2:       "1.2.3",
			expected: true,
		},
		"greater than patch is greater": {
			v1:       "1.2.4",
			v2:       "1.2.3",
			expected: true,
		},
		"less than major is not greater": {
			v1:       "1.2.3",
			v2:       "2.0.0",
			expected: false,
		},
		"less than minor is not greater": {
			v1:       "2.0.3",
			v2:       "2.1.0",
			expected: false,
		},
		"less than patch is not greater": {
			v1:       "2.0.0",
			v2:       "2.0.1",
			expected: false,
		},
		"invalid first version errors": {
			v1:  "dev",
			v2:  "2.0.1",
			err: slackerror.New(slackerror.ErrInvalidSemVer),
		},
		"invalid second version errors": {
			v1:  "2.0.0",
			v2:  "dev",
			err: slackerror.New(slackerror.ErrInvalidSemVer),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := SemVerGreaterThan(tt.v1, tt.v2)
			if tt.err != nil {
				expectedErr := slackerror.ToSlackError(tt.err)
				actualErr := slackerror.ToSlackError(err)
				assert.Equal(t, expectedErr.Code, actualErr.Code)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Update_SemVerLessThan(t *testing.T) {
	tests := map[string]struct {
		v1       string
		v2       string
		expected bool
		err      error
	}{
		"equal versions are not lesser": {
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: false,
		},
		"greater than major is not lesser": {
			v1:       "2.0.0",
			v2:       "1.2.3",
			expected: false,
		},
		"greater than minor is not lesser": {
			v1:       "1.3.0",
			v2:       "1.2.3",
			expected: false,
		},
		"greater than patch is not lesser": {
			v1:       "1.2.4",
			v2:       "1.2.3",
			expected: false,
		},
		"less than major is lesser": {
			v1:       "1.2.3",
			v2:       "2.0.0",
			expected: true,
		},
		"less than minor is lesser": {
			v1:       "2.0.3",
			v2:       "2.1.0",
			expected: true,
		},
		"less than patch is lesser": {
			v1:       "2.0.0",
			v2:       "2.0.1",
			expected: true,
		},
		"invalid first version errors": {
			v1:  "dev",
			v2:  "2.0.1",
			err: slackerror.New(slackerror.ErrInvalidSemVer),
		},
		"invalid second version errors": {
			v1:  "2.0.0",
			v2:  "dev",
			err: slackerror.New(slackerror.ErrInvalidSemVer),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := SemVerLessThan(tt.v1, tt.v2)
			if tt.err != nil {
				expectedErr := slackerror.ToSlackError(tt.err)
				actualErr := slackerror.ToSlackError(err)
				assert.Equal(t, expectedErr.Code, actualErr.Code)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}
