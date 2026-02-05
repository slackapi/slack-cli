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

package slackdeps

import (
	"os"
	"path/filepath"

	"github.com/cli/safeexec"
	"github.com/slackapi/slack-cli/internal/shared/types"
)

// NewOs creates a new SlackOs
func NewOs() *Os {
	return &Os{}
}

// Os contains os package functions that need to be mockable for testing
type Os struct {
	executionDirPathAbs string
}

// Getenv defaults to `os.Getenv` and can be mocked to test
func (c *Os) Getenv(key string) (value string) {
	return os.Getenv(key)
}

// LookPath is a safe alternative to `exec.LookPath()` that
// reduces security risks on Windows and has better PATH support.
func (c *Os) LookPath(file string) (path string, err error) {
	return safeexec.LookPath(file)
}

// LookupEnv defaults to `os.LookupEnv` and can be mocked to test
func (c *Os) LookupEnv(key string) (value string, present bool) {
	return os.LookupEnv(key)
}

// Setenv defaults to `os.Setenv` and can be mocked to test
func (c *Os) Setenv(key string, value string) error {
	return os.Setenv(key, value)
}

// Getwd defaults to `os.Getwd` and can be mocked to test
func (c *Os) Getwd() (dir string, err error) {
	return os.Getwd()
}

// UserHomeDir returns the current user's home directory and can be mocked to test
func (c *Os) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetExecutionDir returns the absolute path where the process started execution
func (c *Os) GetExecutionDir() string {
	return c.executionDirPathAbs
}

// SetExecutionDir sets the absolute path where the process started execution
func (c *Os) SetExecutionDir(dirPathAbs string) {
	c.executionDirPathAbs = dirPathAbs
}

// IsNotExist returns a boolean indicating whether the provided error is known to report that a file or directory does not exist and can be mocked to test
func (c *Os) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
// Reference: https://pkg.go.dev/path/filepath#Glob
func (c *Os) Glob(pattern string) (matches []string, err error) {
	return filepath.Glob(pattern)
}

// Exit exits the program with a return code
func (c *Os) Exit(code int) {
	os.Exit(code)
}

// Stdout returns the file descriptor for stdout
func (c *Os) Stdout() types.File {
	return os.Stdout
}
