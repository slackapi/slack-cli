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

//go:build !windows

package update

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func Test_detach(t *testing.T) {
	command := exec.Command("slack")
	detach(command)
	require.NotNil(t, command.SysProcAttr)
	assert.True(t, command.SysProcAttr.Setsid, "the process starts in a new session")
}

// isDetached returns true when the process leads its own session.
func isDetached() bool {
	sid, err := unix.Getsid(0)
	return err == nil && sid == os.Getpid()
}
