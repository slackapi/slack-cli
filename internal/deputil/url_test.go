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

package deputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_URLChecker(t *testing.T) {
	url := URLChecker("https://github.com/slack-samples/deno-starter-template")
	assert.Equal(t, "https://github.com/slack-samples/deno-starter-template", url, "should return url when url is valid")

	url = URLChecker("fake_url")
	assert.Equal(t, "", url, "should return empty string when url is invalid")
}
