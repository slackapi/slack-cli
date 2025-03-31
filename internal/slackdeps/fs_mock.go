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
	"io/fs"
	"os"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
)

// NewFsMock creates a mock for the file system.
func NewFsMock() *FsMock {
	return &FsMock{}
}

// FsMock is composed of the afero memory-based file system and testify mocking package.
type FsMock struct {
	mock.Mock
	afero.MemMapFs
}

// Mkdir is a hybrid mock function.
// 1. When no mock is defined, then it will call the Afero memory-based file system. This is the most common scenario.
// 2. When a mock is defined, then the mock response is used instead of calling Afero's memory-based file system.
func (m *FsMock) Mkdir(path string, perm fs.FileMode) (_err error) {
	defer func() {
		if err := recover(); err != nil {
			_err = m.MemMapFs.Mkdir(path, perm)
		}
	}()

	args := m.Called(path, perm)
	return args.Error(0)
}

// MkdirAll is a hybrid mock function.
// 1. When no mock is defined, then it will call the Afero memory-based file system. This is the most common scenario.
// 2. When a mock is defined, then the mock response is used instead of calling Afero's memory-based file system.
//
// This is accomplished by catching a runtime panic. When there are no mocks defined, then m.Called() will panic.
// The defer will capture the panic and recover by calling the original memory-based MkdirAll.
// The named return value (_err) is then set at the last moment before being returned to the caller.
func (m *FsMock) MkdirAll(path string, perm fs.FileMode) (_err error) {
	defer func() {
		if err := recover(); err != nil {
			_err = m.MemMapFs.MkdirAll(path, perm)
		}
	}()

	args := m.Called(path, perm)
	return args.Error(0)
}

// OpenFile is a hybrid mock function.
// 1. When no mock is defined, then it will call the Afero memory-based file system. This is the most common scenario.
// 2. When a mock is defined, then the mock response is used instead of calling Afero's memory-based file system.
func (m *FsMock) OpenFile(name string, flag int, perm fs.FileMode) (_file afero.File, _err error) {
	defer func() {
		if err := recover(); err != nil {
			_file, _err = m.MemMapFs.OpenFile(name, flag, perm)
		}
	}()
	args := m.Called(name, flag, perm)
	return args.Get(0).(afero.File), args.Error(1)
}

// Stat is a hybrid mock function.
// 1. When no mock is defined, then it will call the Afero memory-based file system. This is the most common scenario.
// 2. When a mock is defined, then the mock response is used instead of calling Afero's memory-based file system.
func (m *FsMock) Stat(path string) (_info fs.FileInfo, _err error) {
	defer func() {
		if err := recover(); err != nil {
			_info, _err = m.MemMapFs.Stat(path)
		}
	}()

	args := m.Called(path)

	// fs.FileInfo can be nil or defined, so check and set it appropriately
	var fileInfo fs.FileInfo
	if _fileInfo, ok := args.Get(0).(fs.FileInfo); ok {
		fileInfo = _fileInfo
	}

	return fileInfo, args.Error(1)
}

// NewFileMock creates a mock for a file
func NewFileMock() *FileMock {
	return &FileMock{}
}

// FileMock contains mockable information about a file
type FileMock struct {
	FileInfo os.FileInfo
	mock.Mock
}

// Stat returns information about a mocked file
func (m *FileMock) Stat() (os.FileInfo, error) {
	return m.FileInfo, nil
}

// FileInfoNamedPipe holds mocked info of a named pipe file type
type FileInfoNamedPipe struct {
	os.FileInfo
}

// Mode returns the status bits of a named pipe file
func (fi *FileInfoNamedPipe) Mode() os.FileMode {
	return os.ModeNamedPipe
}

// FileInfoCharDevice holds mocked info of a char device file type
type FileInfoCharDevice struct {
	os.FileInfo
}

// Mode returns the status bits of a char device file
func (fi *FileInfoCharDevice) Mode() os.FileMode {
	return os.ModeCharDevice
}
