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
	"fmt"
	"io"

	"github.com/pkg/browser"
)

// Browser contains methods for visiting a webpage
type Browser interface {
	OpenURL(url string)
}

// GoBrowser wraps the browser to handle errors
type GoBrowser struct {
	out io.Writer
}

// NewBrowser prepares a browser for visiting URLs
func NewBrowser(out io.Writer) GoBrowser {
	return GoBrowser{
		out: out,
	}
}

// OpenURL opens the URL in browser or otherwise prints the URL
func (bw GoBrowser) OpenURL(url string) {
	err := browser.OpenURL(url)
	if err != nil {
		_, _ = bw.out.Write([]byte(fmt.Sprintln(url)))
	}
}
