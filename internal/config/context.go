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

package config

import (
	"context"

	"github.com/slackapi/slack-cli/internal/shared/types"
)

type contextKey string

const CONTEXT_ENV contextKey = "env"
const CONTEXT_TOKEN contextKey = "token"
const CONTEXT_TEAM_ID contextKey = "team_id"
const CONTEXT_TEAM_DOMAIN contextKey = "team_domain" // e.g. "subarachnoid"
const CONTEXT_USER_ID contextKey = "user_id"
const CONTEXT_SESSION_ID contextKey = "session_id"
const CONTEXT_ENTERPRISE_ID contextKey = "enterprise_id"
const CONTEXT_TRACE_ID contextKey = "trace_id"

// SetContextApp sets the app on the context
//
// [LEGACY] Please do not use and prefer to
// directly pass the app to methods which require
// an app if possible.
func SetContextApp(ctx context.Context, app types.App) context.Context {
	return context.WithValue(ctx, CONTEXT_ENV, app)
}

// GetContextApp gets an app from the context
//
// [LEGACY] Please do not use and prefer to
// directly pass the app to methods which require
// an app if possible.
func GetContextApp(ctx context.Context) types.App {
	app, ok := ctx.Value(CONTEXT_ENV).(types.App)
	if !ok {
		return types.App{}
	}
	return app
}

func SetContextToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, CONTEXT_TOKEN, token)
}

func GetContextToken(ctx context.Context) string {
	token, ok := ctx.Value(CONTEXT_TOKEN).(string)
	if !ok {
		return ""
	}
	return token
}

func SetContextEnterpriseID(ctx context.Context, enterpriseID string) context.Context {
	return context.WithValue(ctx, CONTEXT_ENTERPRISE_ID, enterpriseID)
}

func GetContextEnterpriseID(ctx context.Context) string {
	enterpriseID, ok := ctx.Value(CONTEXT_ENTERPRISE_ID).(string)
	if !ok {
		return ""
	}
	return enterpriseID
}

func SetContextTeamID(ctx context.Context, teamID string) context.Context {
	return context.WithValue(ctx, CONTEXT_TEAM_ID, teamID)
}
func GetContextTeamID(ctx context.Context) string {
	teamID, ok := ctx.Value(CONTEXT_TEAM_ID).(string)
	if !ok {
		return ""
	}
	return teamID
}

func SetContextTeamDomain(ctx context.Context, teamDomain string) context.Context {
	return context.WithValue(ctx, CONTEXT_TEAM_DOMAIN, teamDomain)
}

func GetContextTeamDomain(ctx context.Context) string {
	teamDomain, ok := ctx.Value(CONTEXT_TEAM_DOMAIN).(string)
	if !ok {
		return ""
	}
	return teamDomain
}

func SetContextUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, CONTEXT_USER_ID, userID)
}

func GetContextUserID(ctx context.Context) string {
	userID, ok := ctx.Value(CONTEXT_USER_ID).(string)
	if !ok {
		return ""
	}
	return userID
}

func GetContextSessionID(ctx context.Context) string {
	sessionID, ok := ctx.Value(CONTEXT_SESSION_ID).(string)
	if !ok {
		return ""
	}
	return sessionID
}

func SetContextSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, CONTEXT_SESSION_ID, sessionID)
}

func SetContextTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, CONTEXT_TRACE_ID, traceID)
}

func GetContextTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(CONTEXT_TRACE_ID).(string)
	if !ok {
		return ""
	}
	return traceID
}
