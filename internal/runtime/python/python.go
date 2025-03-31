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

package python

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

//go:embed hooks.json
var hooksJSON []byte

// slackBoltPackageName is the name of the Bolt for Python dependency in requirements.txt
const slackBoltPackageName = "slack-bolt"

const (
	// slackCLIHooksPackageName the name of the Python Hooks dependency in requirements.txt
	slackCLIHooksPackageName = "slack-cli-hooks"

	// slackCLIHooksPackageVersion is the version range of the Python Hooks dependency in requirements.txt
	slackCLIHooksPackageVersion = "<1.0.0"

	// slackCLIHooksPackageSpecifier is the package name and version range of the Python Hooks dependency in requirements.txt
	slackCLIHooksPackageSpecifier = slackCLIHooksPackageName + slackCLIHooksPackageVersion
)

// Python runtime type
type Python struct {
}

// New creates a new Python runtime
func New() *Python {
	return &Python{}
}

// IgnoreDirectories is a list of directories to ignore when packaging the runtime for deployment.
func (p *Python) IgnoreDirectories() []string {
	return []string{}
}

// InstallProjectDependencies is unsupported by Python because a virtual environment is required before installing the project dependencies.
// TODO(@mbrooks) - should we confirm that the project is using Bolt Python?
func (p *Python) InstallProjectDependencies(ctx context.Context, projectDirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer, fs afero.Fs, os types.Os) (output string, err error) {
	var outputs []string
	var errs []error

	// Defer a function to transform the return values
	defer func() {
		// Manual steps to setup virtual environment and install dependencies
		var activateVirtualEnv = "source .venv/bin/activate"
		if runtime.GOOS == "windows" {
			activateVirtualEnv = `.venv\Scripts\activate`
		}

		// Get the relative path to the project directory
		var projectDirPathRel, _ = getProjectDirRelPath(os, os.GetExecutionDir(), projectDirPath)

		outputs = append(outputs, fmt.Sprintf("Manually setup a %s", style.Highlight("Python virtual environment")))
		if projectDirPathRel != "." {
			outputs = append(outputs, fmt.Sprintf("  Change into the project: %s", style.CommandText(fmt.Sprintf("cd %s%s", filepath.Base(projectDirPathRel), string(filepath.Separator)))))
		}
		outputs = append(outputs, fmt.Sprintf("  Create virtual environment: %s", style.CommandText("python3 -m venv .venv")))
		outputs = append(outputs, fmt.Sprintf("  Activate virtual environment: %s", style.CommandText(activateVirtualEnv)))
		outputs = append(outputs, fmt.Sprintf("  Install project dependencies: %s", style.CommandText("pip install -r requirements.txt")))
		outputs = append(outputs, fmt.Sprintf("  Learn more: %s", style.Underline("https://docs.python.org/3/tutorial/venv.html")))

		// Get first error or nil
		var firstErr error
		if len(errs) > 0 {
			firstErr = errs[0]
		}

		// Update return value
		output = strings.Join(outputs, "\n")
		err = firstErr
	}()

	// Read requirements.txt
	var requirementsFilePath = filepath.Join(projectDirPath, "requirements.txt")

	file, err := afero.ReadFile(fs, requirementsFilePath)
	if err != nil {
		errs = append(errs, err)
		outputs = append(outputs, fmt.Sprintf("Error reading requirements.txt: %s", err))
		return
	}

	fileData := string(file)

	// Skip when slack-cli-hooks is already declared in requirements.txt
	if strings.Contains(fileData, slackCLIHooksPackageName) {
		outputs = append(outputs, fmt.Sprintf("Found requirements.txt with %s", style.Highlight(slackCLIHooksPackageName)))
		return
	}

	// Add slack-cli-hooks to requirements.txt
	//
	// Regex finds all lines that match "slack-bolt" including optional version specifier (e.g. slack-bolt==1.21.2)
	re := regexp.MustCompile(fmt.Sprintf(`(%s.*)`, slackBoltPackageName))
	matches := re.FindAllString(fileData, -1)

	if len(matches) > 0 {
		// Inserted under the slack-bolt dependency
		fileData = re.ReplaceAllString(fileData, fmt.Sprintf("$1\n%s", slackCLIHooksPackageSpecifier))
	} else {
		// Insert at bottom of file
		fileData = fmt.Sprintf("%s\n%s", strings.TrimSpace(fileData), slackCLIHooksPackageSpecifier)
	}

	// Save requirements.txt
	err = afero.WriteFile(fs, requirementsFilePath, []byte(fileData), 0644)
	if err != nil {
		errs = append(errs, err)
		outputs = append(outputs, fmt.Sprintf("Error updating requirements.txt: %s", err))
	} else {
		outputs = append(outputs, fmt.Sprintf("Updated requirements.txt with %s", style.Highlight(slackCLIHooksPackageSpecifier)))
	}

	return
}

// Name prints the name of the runtime
func (p *Python) Name() string {
	return "Python"
}

// Version is the runtime version used by the hosted app deployment
func (p *Python) Version() string {
	return "python" // the server interprets this as: always the latest version layer
}

// SetVersion sets the Version value
func (p *Python) SetVersion(version string) {
	// Unsupported
}

// HooksJSONTemplate returns the default hooks.json template
func (p *Python) HooksJSONTemplate() []byte {
	return hooksJSON
}

// PreparePackage will prepare and copy the app in projectDirPath to tmpDirPath as a release-ready bundle.
func (p *Python) PreparePackage(sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor, opts types.PreparePackageOpts) error {
	return nil // Unsupported
}

// IsRuntimeForProject returns true if dirPath is a Python project
func IsRuntimeForProject(ctx context.Context, fs afero.Fs, dirPath string, sdkConfig hooks.SDKCLIConfig) bool {
	// Is Python project when app manifest says so
	if strings.HasPrefix(sdkConfig.Runtime, "python") {
		return true
	}

	// Python projects must have a requirements.txt in the root dirPath
	var requirementsTxtPath = filepath.Join(dirPath, "requirements.txt")
	if _, err := fs.Stat(requirementsTxtPath); err == nil {
		return true
	}

	return false
}

// getProjectDirRelPath returns the relative path from current working directory
// to projectDirPath or an error. When an error occurs, the projectDirPath is
// returned as the relative path.
func getProjectDirRelPath(os types.Os, currentDirPath string, projectDirPath string) (string, error) {
	// Get the current directory to use as the base for the project
	if currentDirPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return projectDirPath, err
		}
		currentDirPath = cwd
	}
	// Find relative path between current directory and project directory
	filePathRel, err := filepath.Rel(currentDirPath, projectDirPath)
	if err != nil {
		return projectDirPath, err
	}
	return filePathRel, nil
}
