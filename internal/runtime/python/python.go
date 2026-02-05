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

package python

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
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

// installRequirementsTxt handles adding slack-cli-hooks to requirements.txt
func installRequirementsTxt(fs afero.Fs, projectDirPath string) (output string, err error) {
	requirementsFilePath := filepath.Join(projectDirPath, "requirements.txt")

	file, err := afero.ReadFile(fs, requirementsFilePath)
	if err != nil {
		return fmt.Sprintf("Error reading requirements.txt: %s", err), err
	}

	fileData := string(file)

	// Skip when slack-cli-hooks is already declared in requirements.txt
	if strings.Contains(fileData, slackCLIHooksPackageName) {
		return fmt.Sprintf("Found requirements.txt with %s", style.Highlight(slackCLIHooksPackageName)), nil
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
		return fmt.Sprintf("Error updating requirements.txt: %s", err), err
	}

	return fmt.Sprintf("Updated requirements.txt with %s", style.Highlight(slackCLIHooksPackageSpecifier)), nil
}

// installPyProjectToml handles adding slack-cli-hooks to pyproject.toml
func installPyProjectToml(fs afero.Fs, projectDirPath string) (output string, err error) {
	pyProjectFilePath := filepath.Join(projectDirPath, "pyproject.toml")

	file, err := afero.ReadFile(fs, pyProjectFilePath)
	if err != nil {
		return fmt.Sprintf("Error reading pyproject.toml: %s", err), err
	}

	fileData := string(file)

	// Check if slack-cli-hooks is already declared
	if strings.Contains(fileData, slackCLIHooksPackageName) {
		return fmt.Sprintf("Found pyproject.toml with %s", style.Highlight(slackCLIHooksPackageName)), nil
	}

	// Parse only to validate the file structure
	var config map[string]interface{}
	err = toml.Unmarshal(file, &config)
	if err != nil {
		return fmt.Sprintf("Error parsing pyproject.toml: %s", err), err
	}

	// Verify `project` section and `project.dependencies` array exist
	projectSection, exists := config["project"]
	if !exists {
		err := fmt.Errorf("pyproject.toml missing project section")
		return fmt.Sprintf("Error: %s", err), err
	}

	projectMap, ok := projectSection.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("pyproject.toml project section is not a valid format")
		return fmt.Sprintf("Error: %s", err), err
	}

	if _, exists := projectMap["dependencies"]; !exists {
		err := fmt.Errorf("pyproject.toml missing dependencies array")
		return fmt.Sprintf("Error: %s", err), err
	}

	// Use string manipulation to add the dependency while preserving formatting.
	// This regex matches the dependencies array and its contents, handling both single-line and multi-line formats.
	// Note: This regex may not correctly handle commented-out dependencies arrays or nested brackets in string values.
	// These edge cases are uncommon in practice and the TOML validation above will catch malformed files.
	dependenciesRegex := regexp.MustCompile(`(?s)(dependencies\s*=\s*\[)([^\]]*?)(\])`)
	matches := dependenciesRegex.FindStringSubmatch(fileData)

	if len(matches) == 0 {
		err := fmt.Errorf("pyproject.toml missing dependencies array")
		return fmt.Sprintf("Error: %s", err), err
	}

	prefix := matches[1]  // "...dependencies = ["
	content := matches[2] // array contents
	suffix := matches[3]  // "]..."

	// Always append slack-cli-hooks at the end of the dependencies array.
	// Formatting:
	// - Multi-line arrays get a trailing comma to match Python/TOML conventions
	//   and make future additions cleaner.
	// - Single-line arrays omit the trailing comma for a compact appearance,
	//   which is the typical style for short dependency lists.
	var newContent string
	content = strings.TrimRight(content, " \t\n")
	if !strings.HasSuffix(content, ",") {
		content += ","
	}
	if strings.Contains(content, "\n") {
		// Multi-line format: append with proper indentation and trailing comma
		newContent = content + "\n" + `    "` + slackCLIHooksPackageSpecifier + `",` + "\n"
	} else {
		// Single-line format: append inline without trailing comma
		newContent = content + ` "` + slackCLIHooksPackageSpecifier + `"`
	}

	// Replace the dependencies array content
	fileData = dependenciesRegex.ReplaceAllString(fileData, prefix+newContent+suffix)

	// Save pyproject.toml
	err = afero.WriteFile(fs, pyProjectFilePath, []byte(fileData), 0644)
	if err != nil {
		return fmt.Sprintf("Error updating pyproject.toml: %s", err), err
	}

	return fmt.Sprintf("Updated pyproject.toml with %s", style.Highlight(slackCLIHooksPackageSpecifier)), nil
}

// InstallProjectDependencies is unsupported by Python because a virtual environment is required before installing the project dependencies.
// TODO(@mbrooks) - should we confirm that the project is using Bolt Python?
func (p *Python) InstallProjectDependencies(ctx context.Context, projectDirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer, fs afero.Fs, os types.Os) (output string, err error) {
	var outputs []string
	var errs []error

	// Detect which dependency file(s) exist
	requirementsFilePath := filepath.Join(projectDirPath, "requirements.txt")
	pyProjectFilePath := filepath.Join(projectDirPath, "pyproject.toml")

	hasRequirementsTxt := false
	hasPyProjectToml := false

	if _, err := fs.Stat(requirementsFilePath); err == nil {
		hasRequirementsTxt = true
	}

	if _, err := fs.Stat(pyProjectFilePath); err == nil {
		hasPyProjectToml = true
	}

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

		// Provide appropriate install command based on which file exists
		if hasRequirementsTxt {
			outputs = append(outputs, fmt.Sprintf("  Install project dependencies: %s", style.CommandText("pip install -r requirements.txt")))
		}
		if hasPyProjectToml {
			outputs = append(outputs, fmt.Sprintf("  Install project dependencies: %s", style.CommandText("pip install -e .")))
		}

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

	// Handle requirements.txt if it exists
	if hasRequirementsTxt {
		output, err := installRequirementsTxt(fs, projectDirPath)
		outputs = append(outputs, output)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Handle pyproject.toml if it exists
	if hasPyProjectToml {
		output, err := installPyProjectToml(fs, projectDirPath)
		outputs = append(outputs, output)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// If neither file exists, return an error
	if !hasRequirementsTxt && !hasPyProjectToml {
		err := fmt.Errorf("no Python dependency file found (requirements.txt or pyproject.toml)")
		errs = append(errs, err)
		outputs = append(outputs, fmt.Sprintf("Error: %s", err))
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
func (p *Python) PreparePackage(ctx context.Context, sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor, opts types.PreparePackageOpts) error {
	return nil // Unsupported
}

// IsRuntimeForProject returns true if dirPath is a Python project
func IsRuntimeForProject(ctx context.Context, fs afero.Fs, dirPath string, sdkConfig hooks.SDKCLIConfig) bool {
	// Is Python project when app manifest says so
	if strings.HasPrefix(sdkConfig.Runtime, "python") {
		return true
	}

	// Python projects must have a requirements.txt or pyproject.toml in the root dirPath
	var requirementsTxtPath = filepath.Join(dirPath, "requirements.txt")
	if _, err := fs.Stat(requirementsTxtPath); err == nil {
		return true
	}

	var pyProjectTomlPath = filepath.Join(dirPath, "pyproject.toml")
	if _, err := fs.Stat(pyProjectTomlPath); err == nil {
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
