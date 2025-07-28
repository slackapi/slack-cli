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

package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
)

// NPM defines the npm commands to interact with a node project
type NPM interface {
	InstallAllPackages(ctx context.Context, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error)
	InstallDevPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error)
	ListPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (pkgNameVersion string, exists bool)
}

// NPMClient implements the NPM interface and executes real npm commands
type NPMClient struct {
}

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	PackageManager string `json:"packageManager"`
}

// detectPackageManager reads package.json and returns the appropriate package manager command
func detectPackageManager(dirPath string) string {
	packageJSONPath := filepath.Join(dirPath, "package.json")
	
	// Read package.json
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "npm" // Default to npm if can't read package.json
	}
	
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "npm" // Default to npm if can't parse JSON
	}
	
	// Parse packageManager field (e.g., "pnpm@8.0.0" -> "pnpm")
	if pkg.PackageManager != "" {
		parts := strings.Split(pkg.PackageManager, "@")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	
	return "npm" // Default to npm
}

// InstallAllPackages installs all packages by running a command similar to `npm install .`
func (n *NPMClient) InstallAllPackages(ctx context.Context, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error) {
	// Internal hook implementation with a preferred install command
	//
	// Use --no-package-lock to prevent pruning of the node dependency
	//
	// TODO: The SDK should implement this hook instead of hardcoding the command
	//       An internal hook is used for streaming install outputs to debug logs
	
	packageManager := detectPackageManager(dirPath)
	var command string
	
	switch packageManager {
	case "yarn":
		command = "yarn install --verbose"
	case "pnpm":
		command = "pnpm install --reporter=default"
	default:
		command = "npm install --no-package-lock --no-audit --progress=false --loglevel=verbose ."
	}
	
	hookScript := hooks.HookScript{
		Name:    "InstallProjectDependencies",
		Command: command,
	}

	stdout := bytes.Buffer{}

	hookExecOpts := hooks.HookExecOpts{
		Hook:      hookScript,
		Stdin:     ios.ReadIn(),
		Stdout:    &stdout,
		Directory: dirPath,
	}

	_, err := hookExecutor.Execute(ctx, hookExecOpts)
	output := strings.TrimSpace(stdout.String())

	if err != nil {
		ios.PrintDebug(ctx, fmt.Sprintf("Error executing '%s': %s", hookScript.Command, err))
		return "", err
	}

	return output, nil
}

// InstallDevPackage installs the specified package and saves it as a "devDependency"
// by running a command similar to `npm install --save-dev <package>`
func (n *NPMClient) InstallDevPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error) {
	npmSpan, _ := opentracing.StartSpanFromContext(ctx, "npm.install.slack-cli-hooks")
	defer npmSpan.Finish()

	packageManager := detectPackageManager(dirPath)
	var command string
	
	switch packageManager {
	case "yarn":
		command = fmt.Sprintf("yarn add --dev %s", pkgName)
	case "pnpm":
		command = fmt.Sprintf("pnpm add --save-dev %s", pkgName)
	default:
		command = fmt.Sprintf("npm install --save-dev --no-audit --progress=false --loglevel=verbose %s", pkgName)
	}

	hookScript := hooks.HookScript{
		Name:    "InstallProjectDependencies",
		Command: command,
	}

	stdout := bytes.Buffer{}

	hookExecOpts := hooks.HookExecOpts{
		Hook:      hookScript,
		Stdin:     ios.ReadIn(),
		Stdout:    &stdout,
		Directory: dirPath,
	}

	_, err := hookExecutor.Execute(ctx, hookExecOpts)
	output := strings.TrimSpace(stdout.String())

	if err != nil {
		ios.PrintDebug(ctx, fmt.Sprintf("Error executing '%s': %s", hookScript.Command, err))
		npmSpan.SetTag("error", output)
		return "", err
	}

	return output, nil
}

// ListPackage lists the installed version of a provided package
func (n *NPMClient) ListPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (pkgNameVersion string, exists bool) {
	npmSpan, _ := opentracing.StartSpanFromContext(ctx, "npm.list")
	defer npmSpan.Finish()

	packageManager := detectPackageManager(dirPath)
	var command string
	
	switch packageManager {
	case "yarn":
		command = fmt.Sprintf("yarn list --pattern %s --depth=0", pkgName)
	case "pnpm":
		command = fmt.Sprintf("pnpm list %s --depth 0", pkgName)
	default:
		command = fmt.Sprintf("npm list %s --depth 0", pkgName)
	}

	hookScript := hooks.HookScript{
		Name:    "InstallProjectDependencies",
		Command: command,
	}

	stdout := bytes.Buffer{}

	hookExecOpts := hooks.HookExecOpts{
		Hook:      hookScript,
		Stdin:     ios.ReadIn(),
		Stdout:    &stdout,
		Directory: dirPath,
	}

	_, err := hookExecutor.Execute(ctx, hookExecOpts)
	output := strings.TrimSpace(stdout.String())

	if err != nil {
		ios.PrintDebug(ctx, fmt.Sprintf("Error executing '%s': %s", hookScript.Command, err))
		npmSpan.SetTag("error", output)
		return "", false
	}

	// Find the package in the output (e.g. @slack/cli-hooks@1.2.3)
	re := regexp.MustCompile(fmt.Sprintf("%s@.*", pkgName))
	pkgNameVersion = re.FindString(output)
	if strings.TrimSpace(pkgNameVersion) == "" {
		return "", false
	}

	return pkgNameVersion, true
}
