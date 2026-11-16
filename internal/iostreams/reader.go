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

package iostreams

import (
	stdio "io"
	"strings"
)

// Reader contains implementations of a Read methods for various inputs methods
//
// Only stdin is supported for now
type Reader interface {
	ReadIn() stdio.Reader
	ReadInAll() (string, error)
}

// ReadIn returns the reader associated with stdin
func (io *IOStreams) ReadIn() stdio.Reader {
	return io.Stdin
}

// ReadInAll reads all of stdin and returns it with surrounding whitespace trimmed
func (io *IOStreams) ReadInAll() (string, error) {
	raw, err := stdio.ReadAll(io.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}
