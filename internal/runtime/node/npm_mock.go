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

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/stretchr/testify/mock"
)

// NPMMock mocks the NPM interface.
type NPMMock struct {
	mock.Mock
}

// InstallAllPackages mocks the `npm install .` command
func (n *NPMMock) InstallAllPackages(ctx context.Context, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error) {
	args := n.Called(ctx, dirPath, hookExecutor, ios)
	return args.String(0), args.Error(1)
}

// InstallDevPackage mocks the `npm install --save-dev <package>` command
func (n *NPMMock) InstallDevPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (string, error) {
	args := n.Called(ctx, pkgName, dirPath, hookExecutor, ios)
	return args.String(0), args.Error(1)
}

// ListPackage mocks the `npm list` command
func (n *NPMMock) ListPackage(ctx context.Context, pkgName string, dirPath string, hookExecutor hooks.HookExecutor, ios iostreams.IOStreamer) (pkgNameVersion string, exists bool) {
	args := n.Called(ctx, pkgName, dirPath, hookExecutor, ios)
	return args.String(0), args.Bool(1)
}
