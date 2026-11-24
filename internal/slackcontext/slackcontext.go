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

// Package slackcontext defines getters and setters for execution-scoped values
// that are propagated through context.Context during the execution of a Slack
// command. These values may include unique identifiers, session metadata, host
// information, version details, and other execution data.
//
// All values should be set in the context before a Slack command begins execution.
// The values can then be accessed throughout the command's execution lifecycle
// using the provided getter methods.
//
// Each value is stored with an unexported key type to prevent collisions with
// other packages using context values. The package provides type-safe accessors
// for retrieving these values.
package slackcontext

import (
	"context"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// contextKey is an unexported type to avoid context key collisions.
// Learn more: https://github.com/golang/net/blob/236b8f043b920452504e263bc21d354427127473/context/context.go#L100
type contextKey int

const (
	contextKeyProjectID contextKey = iota
	contextKeySessionID
	contextKeySystemID
	contextKeyTraceID
	contextKeyVersion
)

// SetProjectID returns a new `context.Context` that holds a reference to
// the `projectID` and sets the `project_id` attribute on the span in `ctx`.
func SetProjectID(ctx context.Context, projectID string) context.Context {
	// Set projectID in the context
	ctx = context.WithValue(ctx, contextKeyProjectID, projectID)

	trace.SpanFromContext(ctx).SetAttributes(attribute.String("project_id", projectID))

	return ctx
}

// SessionID returns the session ID associated with `ctx`, or
// `""` and `slackerror.ErrContextValueNotFound` if no session ID could be found.
func SessionID(ctx context.Context) (string, error) {
	sessionID, ok := ctx.Value(contextKeySessionID).(string)
	if !ok || sessionID == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for Session ID could not be found")
	}
	return sessionID, nil
}

// SetSessionID returns a new `context.Context` that holds a reference to
// the `sessionID`.
func SetSessionID(ctx context.Context, sessionID string) context.Context {
	ctx = context.WithValue(ctx, contextKeySessionID, sessionID)
	return ctx
}

// SetSystemID returns a new `context.Context` that holds a reference to
// the `systemID` and sets the `system_id` attribute on the span in `ctx`.
func SetSystemID(ctx context.Context, systemID string) context.Context {
	// Set systemID in the context
	ctx = context.WithValue(ctx, contextKeySystemID, systemID)

	trace.SpanFromContext(ctx).SetAttributes(attribute.String("system_id", systemID))

	return ctx
}

func TraceID(ctx context.Context) (string, error) {
	traceID, ok := ctx.Value(contextKeyTraceID).(string)
	if !ok || traceID == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for Trace ID could not be found")
	}
	return traceID, nil
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, contextKeyTraceID, traceID)
	return ctx
}

// Version returns the CLI version associated with `ctx`. If the version is not
// found in the context, it falls back to version.Raw() and returns an error to
// signal that the context was not properly initialized.
func Version(ctx context.Context) (string, error) {
	var v, ok = ctx.Value(contextKeyVersion).(string)
	if !ok || v == "" {
		return version.Raw(), slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for version could not be found in context, falling back to the build version")
	}
	return v, nil
}

// SetVersion adds the slack-cli version to Golang context for trace logging
func SetVersion(ctx context.Context, version string) context.Context {
	ctx = context.WithValue(ctx, contextKeyVersion, version)
	return ctx
}
