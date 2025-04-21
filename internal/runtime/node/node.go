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
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

//go:embed hooks.json
var hooksJSON []byte

// slackCLIHooksPkgName is the name of the npm package for the bolt.js hooks
var slackCLIHooksPkgName = "@slack/cli-hooks"

// Node runtime type
type Node struct {
	npmClient NPM
}

// New creates a new Node runtime
func New() *Node {
	return &Node{
		npmClient: &NPMClient{},
	}
}

// IgnoreDirectories is a list of directories to ignore when packaging the runtime for deployment.
func (n *Node) IgnoreDirectories() []string {
	return []string{"node_modules"}
}

// InstallProjectDependencies runs `npm install` in the dirPath
func (n *Node) InstallProjectDependencies(
	ctx context.Context,
	dirPath string,
	hookExecutor hooks.HookExecutor,
	ios iostreams.IOStreamer,
	fs afero.Fs,
	os types.Os,
) (string, error) {
	var outputs []string
	var errs []error

	// npm list @slack/cli-hooks
	nodePackageNameVersion, exists := n.npmClient.ListPackage(ctx, slackCLIHooksPkgName, dirPath, hookExecutor, ios)
	if exists {
		outputs = append(outputs, fmt.Sprintf("Found package %s", style.Highlight(nodePackageNameVersion)))
	}

	// npm install @slack/cli-hooks
	if !exists {
		_, err := n.npmClient.InstallDevPackage(ctx, slackCLIHooksPkgName, dirPath, hookExecutor, ios)
		nodePackageNameVersion, exists = n.npmClient.ListPackage(ctx, slackCLIHooksPkgName, dirPath, hookExecutor, ios)

		if err != nil || !exists {
			outputs = append(outputs, fmt.Sprintf("Error adding package %s", style.Highlight("@slack/cli-hooks")))
		} else {
			outputs = append(outputs, fmt.Sprintf("Added package %s", style.Highlight(nodePackageNameVersion)))
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	// npm install
	_, err := n.npmClient.InstallAllPackages(ctx, dirPath, hookExecutor, ios)
	if err != nil {
		outputs = append(outputs, fmt.Sprintf("Error installing dependencies using %s", style.Highlight("npm install")))
		errs = append(errs, err)
	} else {
		outputs = append(outputs, fmt.Sprintf("Installed dependencies using %s", style.Highlight("npm install")))
	}

	// Get first error or nil
	var firstErr error
	if len(errs) > 0 {
		firstErr = errs[0]
	}

	return strings.Join(outputs, "\n"), firstErr
}

// Name prints the name of the runtime
func (n *Node) Name() string {
	return "Node.js"
}

// Version is the runtime version used by the hosted app deployment
func (n *Node) Version() string {
	return "node" // the server interprets this as: always the latest version layer
}

// SetVersion sets the Version value
func (n *Node) SetVersion(version string) {
	// Unsupported
}

// HooksJSONTemplate returns the default hooks.json template
func (n *Node) HooksJSONTemplate() []byte {
	return hooksJSON
}

// PreparePackage will prepare and copy the app in projectDirPath to tmpDirPath as a release-ready bundle.
func (n *Node) PreparePackage(ctx context.Context, sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor, opts types.PreparePackageOpts) error {
	return nil // Unsupported
}

// IsRuntimeForProject returns true if dirPath is a Node.js project
func IsRuntimeForProject(ctx context.Context, fs afero.Fs, dirPath string, sdkConfig hooks.SDKCLIConfig) bool {
	// Is Node.js project when app manifest says so
	if strings.HasPrefix(sdkConfig.Runtime, "node") {
		return true
	}

	// Node.js projects must have a package.json in the root dirPath
	var packageJSONPath = filepath.Join(dirPath, "package.json")
	if _, err := fs.Stat(packageJSONPath); err == nil {
		return true
	}
	return false
}
