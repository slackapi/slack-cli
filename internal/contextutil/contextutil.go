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

package contextutil

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

type ctxKey string

const (
	TracerCtxKey  ctxKey = "__TRACER__"
	SpanCtxKey    ctxKey = "__SPAN__"
	VersionCtxKey ctxKey = "__VERSION__"
)

// AddTracerToContext adds the given tracer to request context
func AddTracerToContext(ctx context.Context, tracer opentracing.Tracer) context.Context {
	return context.WithValue(ctx, TracerCtxKey, tracer)
}

// AddVersionToContext adds the slack-cli version to request context
func AddVersionToContext(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, VersionCtxKey, version)
}

// AddSpanToContext adds the given span to request context
func AddSpanToContext(ctx context.Context, span opentracing.Span) context.Context {
	return context.WithValue(ctx, SpanCtxKey, span)
}

// VersionFromContext get the slack-cli version to request context
func VersionFromContext(ctx context.Context) string {
	var version, ok = ctx.Value(VersionCtxKey).(string)
	if !ok {
		version = ""
	}
	return version
}
