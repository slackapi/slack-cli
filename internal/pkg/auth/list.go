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

package auth

import (
	"context"
	"sort"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// List returns a list of the authenticated Slack accounts.
func List(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger) ([]types.SlackAuth, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pkg.auth.list")
	defer span.Finish()

	// Get a sorted list of userAuths
	auths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return []types.SlackAuth{}, slackerror.Wrap(err, "Failed to get a list of user authorizations")
	}

	sort.SliceStable(auths, func(i, j int) bool {
		return auths[i].TeamDomain < auths[j].TeamDomain
	})

	// Notify listeners
	log.Data = logger.LogData{}
	log.Data["userAuthList"] = auths
	log.Log("info", "on_auth_list")

	return auths, nil
}
