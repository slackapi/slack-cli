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

package deno

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

//go:embed hooks.json
var hooksJSON []byte

// Deno runtime type
type Deno struct {
	version string // Defines the hosted deployment runtime and version: "deno" is the latest, "deno1.1", "deno1.x", etc
}

// Constants
const defaultVersion = "deno" // latest version of deno

// New creates a new Deno runtime
func New() *Deno {
	return &Deno{
		version: defaultVersion,
	}
}

// IgnoreDirectories is a list of directories to ignore when packaging the runtime for deployment.
func (d *Deno) IgnoreDirectories() []string {
	return []string{}
}

// InstallProjectDependencies for the Deno project
func (d *Deno) InstallProjectDependencies(
	ctx context.Context,
	dirPath string,
	hookExecutor hooks.HookExecutor,
	ios iostreams.IOStreamer,
	fs afero.Fs,
	os types.Os,
) (
	string,
	error,
) {
	// Check that the deno runtime is installed on the system
	if _, err := os.LookPath("deno"); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrDenoNotFound)
	}

	// Cache the missing Deno app dependencies
	// Ignore error because we don't want errors while caching dependencies to block a task
	_, _ = cacheDenoDependencies(ctx, fs, ios, hookExecutor, dirPath)

	response := fmt.Sprintf("Cached dependencies with %s", style.Darken("deno cache import_map.json"))
	return response, nil
}

// Name prints the name of the runtime
func (d *Deno) Name() string {
	return "Deno"
}

// Version is the runtime version used by the hosted app deployment
func (d *Deno) Version() string {
	if d.version == "" {
		d.version = defaultVersion
	}
	return d.version
}

// SetVersion sets the Version value
func (d *Deno) SetVersion(version string) {
	d.version = version
}

// HooksJSONTemplate returns the default hooks.json template
func (d *Deno) HooksJSONTemplate() []byte {
	return hooksJSON
}

// PreparePackage will prepare and copy the app in srcDirPath to dstDirPath as a release-ready bundle.
func (d *Deno) PreparePackage(ctx context.Context, sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor, opts types.PreparePackageOpts) error {
	// Generate the bundle.js in the dstDirPath
	var packageHookOpts = hooks.HookExecOpts{
		Directory: opts.SrcDirPath,
		Args: map[string]string{
			"source": opts.SrcDirPath,
			"output": opts.DstDirPath,
		},
		Env: map[string]string{
			"DENO_AUTH_TOKENS": opts.AuthTokens,
		},
		Hook: sdkConfig.Hooks.BuildProject,
	}

	// Execute the package hook and ignore the output because it's always 0 length
	_, err := hookExecutor.Execute(ctx, packageHookOpts)
	if err != nil {
		return err
	}

	return nil
}

// IsRuntimeForProject returns true if dirPath is a Deno project
func IsRuntimeForProject(ctx context.Context, fs afero.Fs, dirPath string, sdkConfig hooks.SDKCLIConfig) bool {
	// Is Deno project when app manifest says so
	if strings.HasPrefix(sdkConfig.Runtime, "deno") {
		return true
	}

	// Project files unique to Deno
	var filePaths = []string{
		filepath.Join(dirPath, "deno.json"),
		filepath.Join(dirPath, "deno.jsonc"),
		filepath.Join(dirPath, "import_map.json"),
	}

	for _, filePath := range filePaths {
		if _, err := fs.Stat(filePath); err == nil {
			return true
		}
	}

	return false
}

// cacheDenoDependencies will attempt to download the Deno dependencies from the project in dirPath.
func cacheDenoDependencies(
	ctx context.Context,
	fs afero.Fs,
	ios iostreams.IOStreamer,
	hookExecutor hooks.HookExecutor,
	dirPath string,
) (
	string,
	error,
) {
	stdout := bytes.Buffer{}

	// Run deno cache on each file entry point that's supported by the Deno Slack SDK
	var filenames = []string{"manifest.ts", "manifest.js"}
	for _, filename := range filenames {
		// Skip when the entry point filename doesn't exist
		filePath := filepath.Join(dirPath, filename)
		if _, err := fs.Stat(filePath); err != nil {
			continue
		}

		// Default command args: `deno cache manifest.ts`
		// Skip `--reload` to only download missing dependencies, which is quicker
		cmdArgs := []string{"cache", filePath}

		// Optional import map arg: `deno cache manifest.ts --import-map import_map.json`
		importMapFilePath := filepath.Join(dirPath, "import_map.json")
		if _, err := fs.Stat(importMapFilePath); !os.IsNotExist(err) {
			cmdArgs = append(cmdArgs, "--import-map", importMapFilePath)
		}

		// Internal hook implementation with a preferred cache command
		//
		// TODO: The SDK should implement this hook instead of hardcoding the command
		//       An internal hook is used for streaming install outputs to debug logs
		hookScript := hooks.HookScript{
			Name:    "InstallProjectDependencies",
			Command: fmt.Sprintf("deno %s", strings.Join(cmdArgs, " ")),
		}
		hookExecOpts := hooks.HookExecOpts{
			Hook:   hookScript,
			Stdin:  ios.ReadIn(),
			Stdout: &stdout,
		}

		if _, err := hookExecutor.Execute(ctx, hookExecOpts); err != nil {
			ios.PrintDebug(ctx, "failed to cache project dependencies")
			return "", err
		}
	}
	return strings.TrimSpace(stdout.String()), nil
}
