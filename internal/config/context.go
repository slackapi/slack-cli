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
)

type contextKey string

const ContextToken contextKey = "token"
const contextTeamID contextKey = "team_id"
const contextTeamDomain contextKey = "team_domain" // e.g. "subarachnoid"
const contextUserID contextKey = "user_id"
const contextEnterpriseID contextKey = "enterprise_id"

func SetContextToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ContextToken, token)
}

func GetContextToken(ctx context.Context) string {
	token, ok := ctx.Value(ContextToken).(string)
	if !ok {
		return ""
	}
	return token
}

func SetContextEnterpriseID(ctx context.Context, enterpriseID string) context.Context {
	return context.WithValue(ctx, contextEnterpriseID, enterpriseID)
}

func GetContextEnterpriseID(ctx context.Context) string {
	enterpriseID, ok := ctx.Value(contextEnterpriseID).(string)
	if !ok {
		return ""
	}
	return enterpriseID
}

func SetContextTeamID(ctx context.Context, teamID string) context.Context {
	return context.WithValue(ctx, contextTeamID, teamID)
}
func GetContextTeamID(ctx context.Context) string {
	teamID, ok := ctx.Value(contextTeamID).(string)
	if !ok {
		return ""
	}
	return teamID
}

func SetContextTeamDomain(ctx context.Context, teamDomain string) context.Context {
	return context.WithValue(ctx, contextTeamDomain, teamDomain)
}

func GetContextTeamDomain(ctx context.Context) string {
	teamDomain, ok := ctx.Value(contextTeamDomain).(string)
	if !ok {
		return ""
	}
	return teamDomain
}

func SetContextUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextUserID, userID)
}

func GetContextUserID(ctx context.Context) string {
	userID, ok := ctx.Value(contextUserID).(string)
	if !ok {
		return ""
	}
	return userID
}
