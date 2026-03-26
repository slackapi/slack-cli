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

package hooks

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/spf13/afero"
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
	GetProcess() *os.Process
}

type ShellExec struct{}

// command sets arguments for the shell supported by the current operating system
func (ShellExec) command(name string, arg ...string) *exec.Cmd {
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

// GetProcess returns the underlying process
func (e execCommander) GetProcess() *os.Process {
	if e.Cmd != nil {
		return e.Process
	}
	return nil
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

// ShellEnv builds the environment variables for a hook command.
func (opts HookExecOpts) ShellEnv(ctx context.Context, fs afero.Fs, io iostreams.IOStreamer) []string {
	// Gather environment variables saved to the project ".env" file
	dotEnv, err := slackdotenv.Read(fs)
	if err != nil {
		io.PrintDebug(ctx, "Warning: failed to parse .env file: %s", err)
	}
	if len(dotEnv) > 0 {
		keys := make([]string, 0, len(dotEnv))
		for k := range dotEnv {
			keys = append(keys, k)
		}
		io.PrintDebug(ctx, "Loaded variables from .env file: %s", strings.Join(keys, ", "))
	}

	// Whatever cmd.Env is set to will be the ONLY environment variables that the `cmd` will have access to when it runs.
	//
	// Order of precedence from lowest to highest:
	// 1. Provided "opts.Env" variables
	// 2. Saved ".env" file
	// 3. Existing shell environment
	//
	// > Each entry is of the form "key=value".
	// > ...
	// > If Env contains duplicate environment keys, only the last value in the slice for each duplicate key is used.
	//
	// https://pkg.go.dev/os/exec#Cmd.Env
	var cmdEnvVars []string
	for name, value := range opts.Env {
		cmdEnvVars = append(cmdEnvVars, name+"="+value)
	}
	for k, v := range dotEnv {
		cmdEnvVars = append(cmdEnvVars, k+"="+v)
	}
	cmdEnvVars = append(cmdEnvVars, os.Environ()...)
	return cmdEnvVars
}
