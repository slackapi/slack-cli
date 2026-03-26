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
	"os"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/spf13/afero"
)

type HookExecutor interface {
	Execute(ctx context.Context, opts HookExecOpts) (response string, err error)
}

func GetHookExecutor(ios iostreams.IOStreamer, fs afero.Fs, cfg SDKCLIConfig) HookExecutor {
	protocol := cfg.Config.SupportedProtocols.Preferred()
	switch protocol {
	case HookProtocolV2:
		return &HookExecutorMessageBoundaryProtocol{
			IO: ios,
			Fs: fs,
		}
	default:
		return &HookExecutorDefaultProtocol{
			IO: ios,
			Fs: fs,
		}
	}
}

func processExecOpts(ctx context.Context, opts HookExecOpts, fs afero.Fs, io iostreams.IOStreamer) ([]string, []string, []string, error) {
	cmdStr, err := opts.Hook.Get()
	if err != nil {
		return []string{}, []string{}, []string{}, err
	}

	// We're taking the script and separating it into individual fields to be compatible with Exec.Command,
	// then appending any additional arguments as flag --key=value pairs.
	cmdArgs := strings.Fields(cmdStr)
	var cmdArgVars = cmdArgs[1:] // omit the first item because that is the command name
	cmdArgVars = append(cmdArgVars, goutils.MapToStringSlice(opts.Args, "--")...)

	// Load .env file variables
	dotEnv, err := slackdotenv.Read(fs)
	if err != nil {
		io.PrintDebug(ctx, "Warning: failed to parse .env file: %s", err)
	}
	if len(dotEnv) > 0 {
		keys := make([]string, 0, len(dotEnv))
		for k := range dotEnv {
			keys = append(keys, k)
		}
		io.PrintDebug(ctx, "loaded variables from .env file: %s", strings.Join(keys, ", "))
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

	return cmdArgs, cmdArgVars, cmdEnvVars, nil
}
