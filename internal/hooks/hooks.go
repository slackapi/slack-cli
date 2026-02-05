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
)

type HookExecutor interface {
	Execute(ctx context.Context, opts HookExecOpts) (response string, err error)
}

func GetHookExecutor(ios iostreams.IOStreamer, cfg SDKCLIConfig) HookExecutor {
	protocol := cfg.Config.SupportedProtocols.Preferred()
	switch protocol {
	case HookProtocolV2:
		return &HookExecutorMessageBoundaryProtocol{
			IO: ios,
		}
	default:
		return &HookExecutorDefaultProtocol{
			IO: ios,
		}
	}
}

func processExecOpts(opts HookExecOpts) ([]string, []string, []string, error) {
	cmdStr, err := opts.Hook.Get()
	if err != nil {
		return []string{}, []string{}, []string{}, err
	}

	// We're taking the script and separating it into individual fields to be compatible with Exec.Command,
	// then appending any additional arguments as flag --key=value pairs.
	cmdArgs := strings.Fields(cmdStr)
	var cmdArgVars = cmdArgs[1:] // omit the first item because that is the command name
	cmdArgVars = append(cmdArgVars, goutils.MapToStringSlice(opts.Args, "--")...)

	// Whatever cmd.Env is set to will be the ONLY environment variables that the `cmd` will have access to when it runs.
	// To avoid removing any environment variables that are set in the current environment, we first set the cmd.Env to the current environment.
	// before adding any new environment variables.
	var cmdEnvVars = os.Environ()
	cmdEnvVars = append(cmdEnvVars, goutils.MapToStringSlice(opts.Env, "")...)

	return cmdArgs, cmdArgVars, cmdEnvVars, nil
}
