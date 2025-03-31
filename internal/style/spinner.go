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

package style

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
)

// isSpinAllowed decides if the spinner should animate or print on changes
var isSpinAllowed bool

// Spinner is a stylized object to indicate loading text
type Spinner struct {
	spinner *spinner.Spinner
	text    string
	icon    string

	active bool      // active track if the spinner is between start and stop
	writer io.Writer // writer outputs updates to certain location
}

// NewSpinner creates a spinner that will write to writer
func NewSpinner(writer io.Writer) *Spinner {
	return &Spinner{
		spinner: spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(writer)),
		text:    "",
		icon:    "",
		writer:  writer,
	}
}

// ToggleSpinner enables the animation when spins is true, otherwise disables it
func ToggleSpinner(spins bool) {
	isSpinAllowed = spins
}

// Start makes the spinner active with prepared text
func (s *Spinner) Start() {
	s.spinner.Suffix = " " + s.text
	s.active = true
	if isSpinAllowed {
		s.spinner.Start()
	} else {
		_, _ = s.writer.Write([]byte(s.icon + s.text + "\n"))
	}
}

// Stop halts the spinning and displays completion text
func (s *Spinner) Stop() {
	s.spinner.FinalMSG = fmt.Sprintf(
		"%s%s\n",
		s.icon,
		s.text,
	)
	s.active = false
	s.spinner.Stop()
}

// Active is true if the spinner is started
func (s *Spinner) Active() bool {
	return s.active
}

// Status returns the current spinner text and spin status
func (s *Spinner) Status() (string, bool) {
	return s.text, s.Active()
}

// Update replaces the spinner text and icon
func (s *Spinner) Update(text string, icon string) *Spinner {
	s.text = text
	s.icon = Emoji(icon)
	if !isSpinAllowed && s.active {
		_, _ = s.writer.Write([]byte(s.icon + s.text + "\n"))
	}
	return s
}
