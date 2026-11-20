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

package tracer

import (
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
)

func Test_SetTraceHeaders(t *testing.T) {
	tracer, closer := jaeger.NewTracer("test", jaeger.NewConstSampler(true), jaeger.NewNullReporter())
	defer closer.Close()
	parent := tracer.StartSpan("parent")
	span := tracer.StartSpan("child", opentracing.ChildOf(parent.Context()))
	header := http.Header{}
	SetTraceHeaders(span, header)
	spanContext := span.Context().(jaeger.SpanContext)
	assert.Equal(t, "0", header.Get("x-b3-sampled"))
	assert.Equal(t, spanContext.SpanID().String(), header.Get("x-b3-spanid"))
	assert.Equal(t, spanContext.TraceID().String(), header.Get("x-b3-traceid"))
	assert.Equal(t, parent.Context().(jaeger.SpanContext).SpanID().String(), header.Get("x-b3-parentspanid"))

	noopHeader := http.Header{}
	SetTraceHeaders(opentracing.NoopTracer{}.StartSpan("noop"), noopHeader)
	assert.Empty(t, noopHeader)
}

func Test_TraceID(t *testing.T) {
	tracer, closer := jaeger.NewTracer("test", jaeger.NewConstSampler(true), jaeger.NewNullReporter())
	defer closer.Close()
	span := tracer.StartSpan("span")
	assert.Equal(t, span.Context().(jaeger.SpanContext).TraceID().String(), TraceID(span))
	assert.Empty(t, TraceID(opentracing.NoopTracer{}.StartSpan("noop")))
}
