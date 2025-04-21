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

package runtime

import (
	"context"
	"strings"

	"github.com/slackapi/slack-cli/internal/deputil"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/runtime/deno"
	"github.com/slackapi/slack-cli/internal/runtime/node"
	"github.com/slackapi/slack-cli/internal/runtime/python"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// DefaultRuntimeName is the runtime used when none are specified or detected
const DefaultRuntimeName = "deno"

// Runtime interface generalizes the required functionality for any runtime language.
// Example runtimes: Node.js, Deno, Python, Java, etc.
type Runtime interface {
	IgnoreDirectories() []string
	InstallProjectDependencies(context.Context, string, hooks.HookExecutor, iostreams.IOStreamer, afero.Fs, types.Os) (string, error)
	Name() string
	Version() string
	SetVersion(string)
	HooksJSONTemplate() []byte
	PreparePackage(context.Context, hooks.SDKCLIConfig, hooks.HookExecutor, types.PreparePackageOpts) error
}

// New creates a new runtime using runtimeName to choose the runtime
//
// runtimeName uses the hosted version pattern, e.g. deno1.1, deno1.x
func New(runtimeName string) (Runtime, error) {
	var rt Runtime
	switch {
	case strings.HasPrefix(strings.ToLower(runtimeName), "deno"): // deno, deno1.x
		rt = deno.New()
		rt.SetVersion(deputil.GetDenoVersion())
	case strings.HasPrefix(strings.ToLower(runtimeName), "node"): // node
		rt = node.New()
		rt.SetVersion(runtimeName) // Default to runtimeName for now
	case strings.HasPrefix(strings.ToLower(runtimeName), "python"): // python
		rt = python.New()
		rt.SetVersion(runtimeName) // Default to runtimeName for now
	default:
		return nil, slackerror.New(slackerror.ErrRuntimeNotSupported).
			WithMessage("This CLI does not support the '%s' runtime", runtimeName)
	}
	return rt, nil
}

// NewDetectProject returns a new Runtime based on the project type of dirPath
func NewDetectProject(ctx context.Context, fs afero.Fs, dirPath string, sdkConfig hooks.SDKCLIConfig) (Runtime, error) {
	var rt Runtime
	var err error
	switch {
	case deno.IsRuntimeForProject(ctx, fs, dirPath, sdkConfig):
		rt, err = New("deno")
	case node.IsRuntimeForProject(ctx, fs, dirPath, sdkConfig):
		rt, err = New("node")
	case python.IsRuntimeForProject(ctx, fs, dirPath, sdkConfig):
		rt, err = New("python")
	default:
		return nil, slackerror.New(slackerror.ErrRuntimeNotSupported).
			WithMessage("Failed to detect project runtime from directory structure")
	}
	return rt, err
}
