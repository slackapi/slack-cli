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

//go:build windows

package update

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func Test_detach(t *testing.T) {
	command := exec.Command("slack")
	detach(command)
	require.NotNil(t, command.SysProcAttr)
	assert.Equal(t, uint32(windows.CREATE_NO_WINDOW|windows.CREATE_NEW_PROCESS_GROUP), command.SysProcAttr.CreationFlags)
}

// isDetached returns true when the process has no console window.
func isDetached() bool {
	window, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	return window == 0
}
