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

package hooks

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ExecInterface is an interface for running shell commands in the OS
type ExecInterface interface {
	Command(env []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, name string, arg ...string) ShellCommand
}

type ShellCommand interface {
	Run() error
	CombinedOutput() ([]byte, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

type ShellExec struct{}

// command sets arguments for the shell supported by the current operating system
func (_ ShellExec) command(name string, arg ...string) *exec.Cmd {
	script := fmt.Sprintf("%s %s", name, strings.Join(arg, " "))
	switch {
	case runtime.GOOS == "windows":
		return exec.Command("powershell", "-Command", script)
	case os.Getenv("SHELL") != "":
		return exec.Command(os.Getenv("SHELL"), "-c", script)
	default:
		return exec.Command("sh", "-c", script)
	}
}

// Command creates a command ready to be run with the current processes shell
func (sh ShellExec) Command(env []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, name string, arg ...string) ShellCommand {
	cmd := sh.command(name, arg...)
	cmd.Env = env
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	return execCommander{cmd}
}

// execCommander wraps the command value from the exec package
type execCommander struct {
	*exec.Cmd
}

type HookExecOpts struct {
	Directory string
	Hook      HookScript
	Args      map[string]string
	Env       map[string]string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Exec      ExecInterface
}
