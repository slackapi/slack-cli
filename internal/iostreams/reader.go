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

package iostreams

import (
	"io"
)

// Reader contains implmentations of a Read methods for various inputs methods
//
// Only stdin is supported for now
type Reader interface {
	ReadIn() io.Reader
}

// ReadIn returns the reader associated with stdin
func (io *IOStreams) ReadIn() io.Reader {
	return io.Stdin
}
