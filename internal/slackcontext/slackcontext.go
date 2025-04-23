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

// Package slackcontext defines getters and setters for request-scoped values
// that are propagated through context.Context during the execution of a Slack
// command. These values may include identifiers, metadata, host information,
// and other execution data.
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

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// contextKey is an unexported type to avoid context key collisions.
// Learn more: https://github.com/golang/net/blob/236b8f043b920452504e263bc21d354427127473/context/context.go#L100
type contextKey int

const (
	contextKeyOpenTracingTraceID contextKey = iota
	contextKeyOpenTracingTracer
	contextKeyProjectID
	contextKeySessionID
	contextKeySystemID
	contextKeyVersion
)

// OpenTracingSpan returns the `opentracing.Span“ associated with `ctx`, or
// `nil` and `slackerror.ErrContextValueNotFound` if no `Span` could be found.
func OpenTracingSpan(ctx context.Context) (opentracing.Span, error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil, slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for OpenTracing Span could not be found")
	}
	return span, nil
}

// SetOpenTracingSpan returns a new `context.Context` that holds a reference to
// the span. If span is nil, a new context without an active span is returned.
func SetOpenTracingSpan(ctx context.Context, span opentracing.Span) context.Context {
	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx
}

// OpenTracingTraceID returns the trace ID associated with `ctx`, or
// `""` and `slackerror.ErrContextValueNotFound` if no trace ID could be found.
func OpenTracingTraceID(ctx context.Context) (string, error) {
	traceID, ok := ctx.Value(contextKeyOpenTracingTraceID).(string)
	if !ok || traceID == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for OpenTracing Trace ID could not be found")
	}
	return traceID, nil
}

// SetOpenTracingTraceID returns a new `context.Context` that holds a reference to
// the `traceID`.
func SetOpenTracingTraceID(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, contextKeyOpenTracingTraceID, traceID)
	return ctx
}

// OpenTracingTracer returns the `opentracing.Tracer` associated with `ctx`, or
// `nil` and `slackerror.ErrContextValueNotFound` if no `Tracer` could be found.
func OpenTracingTracer(ctx context.Context) (opentracing.Tracer, error) {
	var tracer, ok = ctx.Value(contextKeyOpenTracingTracer).(opentracing.Tracer)
	if !ok || tracer == nil {
		return nil, slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for OpenTracing Tracer could not be found")
	}
	return tracer, nil
}

// SetOpenTracingSpan returns a new `context.Context` that holds a reference to
// the `opentracing.Tracer`. If tracer is nil, a new context without a tracer returned.
func SetOpenTracingTracer(ctx context.Context, tracer opentracing.Tracer) context.Context {
	ctx = context.WithValue(ctx, contextKeyOpenTracingTracer, tracer)
	return ctx
}

// ProjectID returns the project ID associated with `ctx`, or
// `""` and `slackerror.ErrContextValueNotFound` if no project ID could be found.
func ProjectID(ctx context.Context) (string, error) {
	var projectID, ok = ctx.Value(contextKeyProjectID).(string)
	if !ok || projectID == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for Project ID could not be found")
	}
	return projectID, nil
}

// SetProjectID returns a new `context.Context` that holds a reference to
// the `projectID` and updates the `opentracing.Span` tag with the `projectID`.
func SetProjectID(ctx context.Context, projectID string) context.Context {
	// Set projectID in the context
	ctx = context.WithValue(ctx, contextKeyProjectID, projectID)

	// Set projectID in OpenTracing by extracting the span and updating the tag
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.SetTag("project_id", projectID)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

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

// SystemID returns the session ID associated with `ctx`, or
// `""` and `slackerror.ErrContextValueNotFound` if no system ID could be found.
func SystemID(ctx context.Context) (string, error) {
	var systemID, ok = ctx.Value(contextKeySystemID).(string)
	if !ok || systemID == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for System ID could not be found")
	}
	return systemID, nil
}

// SetSystemID returns a new `context.Context` that holds a reference to
// the `systemID` and updates the `opentracing.Span` tag with the `systemID`.
func SetSystemID(ctx context.Context, systemID string) context.Context {
	// Set systemID in the context
	ctx = context.WithValue(ctx, contextKeySystemID, systemID)

	// Set projectID in OpenTracing by extracting the span and updating the tag
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.SetTag("system_id", systemID)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return ctx
}

// Version returns the CLI version associated with `ctx`, or
// `""` and `slackerror.ErrContextValueNotFound` if no version could be found.
func Version(ctx context.Context) (string, error) {
	var version, ok = ctx.Value(contextKeyVersion).(string)
	if !ok || version == "" {
		return "", slackerror.New(slackerror.ErrContextValueNotFound).
			WithMessage("The value for Version could not be found")
	}
	return version, nil
}

// SetVersion adds the slack-cli version to Golang context for trace logging
func SetVersion(ctx context.Context, version string) context.Context {
	ctx = context.WithValue(ctx, contextKeyVersion, version)
	return ctx
}
