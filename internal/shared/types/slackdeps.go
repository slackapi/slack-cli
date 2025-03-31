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

package types

import (
	"os"
)

// File represents an open file descriptor
type File interface {
	// Stat returns the FileInfo that describes a file
	Stat() (os.FileInfo, error)
}

// Os are a group of operating system functions following the `os` interface that are shared by all packages and enables testing & mocking
type Os interface {
	// Getenv defaults to `os.Getenv` and can be mocked to test
	Getenv(key string) (value string)

	// LookPath defaults to `os.LookPath` and can be mocked to test
	LookPath(file string) (path string, err error)

	// LookupEnv defaults to `os.LookupEnv` and can be mocked to test
	LookupEnv(key string) (value string, present bool)

	// Setenv defaults to `os.Setenv` and can be mocked to test
	Setenv(key string, value string) error

	// Getwd defaults to `os.Getwd` and can be mocked to test
	Getwd() (dir string, err error)

	// UserHomeDir returns the current user's home directory and can be mocked to test
	UserHomeDir() (dir string, err error)

	// GetExecutionDir returns the absolute path where the process started execution
	GetExecutionDir() string

	// SetExecutionDir sets the absolute path where the process started execution
	SetExecutionDir(dirPathAbs string)

	// IsNotExist returns a boolean indicating whether the provided error is known to report that a file or directory does not exist
	IsNotExist(error) bool

	// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
	// Reference: https://pkg.go.dev/path/filepath#Glob
	Glob(pattern string) (matches []string, err error)

	// Exit causes the program to exit and return with the status code
	Exit(code int)

	// Stdout returns the file descriptor for stdout
	Stdout() File
}
