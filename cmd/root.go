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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	apicmd "github.com/slackapi/slack-cli/cmd/api"
	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/cmd/auth"
	"github.com/slackapi/slack-cli/cmd/blocks"
	"github.com/slackapi/slack-cli/cmd/collaborators"
	"github.com/slackapi/slack-cli/cmd/datastore"
	"github.com/slackapi/slack-cli/cmd/docgen"
	"github.com/slackapi/slack-cli/cmd/docs"
	"github.com/slackapi/slack-cli/cmd/doctor"
	"github.com/slackapi/slack-cli/cmd/env"
	"github.com/slackapi/slack-cli/cmd/externalauth"
	"github.com/slackapi/slack-cli/cmd/feedback"
	"github.com/slackapi/slack-cli/cmd/fingerprint"
	"github.com/slackapi/slack-cli/cmd/function"
	"github.com/slackapi/slack-cli/cmd/help"
	issueresolvercmd "github.com/slackapi/slack-cli/cmd/issueresolver"
	"github.com/slackapi/slack-cli/cmd/manifest"
	"github.com/slackapi/slack-cli/cmd/openformresponse"
	"github.com/slackapi/slack-cli/cmd/platform"
	"github.com/slackapi/slack-cli/cmd/project"
	"github.com/slackapi/slack-cli/cmd/sandbox"
	"github.com/slackapi/slack-cli/cmd/triggers"
	"github.com/slackapi/slack-cli/cmd/upgrade"
	versioncmd "github.com/slackapi/slack-cli/cmd/version"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/internal/update"
	"github.com/slackapi/slack-cli/internal/useragent"
	"github.com/slackapi/slack-cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)
