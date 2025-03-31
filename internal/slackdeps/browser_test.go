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

package slackdeps

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserOpenURL(t *testing.T) {
	t.Run("print the url if opening the browser fails", func(t *testing.T) {
		path := os.Getenv("PATH")
		err := os.Setenv("PATH", "")
		require.NoError(t, err)
		defer os.Setenv("PATH", path)
		buff := bytes.Buffer{}
		logs := log.New(&buff, "", 0)
		browser := NewBrowser(logs.Writer())
		browser.OpenURL("https://example.com")
		assert.Equal(t, buff.String(), "https://example.com\n")
	})
}
