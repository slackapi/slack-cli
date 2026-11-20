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
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func Test_SetTraceHeaders(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tracer := newBridgeTracer(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	span := tracer.StartSpan("child", opentracing.ChildOf(tracer.StartSpan("parent").Context()))
	header := http.Header{}
	SetTraceHeaders(span, header)
	span.Finish()
	spanContext := recorder.Ended()[0].SpanContext()
	assert.Equal(t, "0", header.Get("x-b3-sampled"))
	assert.Equal(t, spanContext.SpanID().String(), header.Get("x-b3-spanid"))
	assert.Equal(t, spanContext.TraceID().String(), header.Get("x-b3-traceid"))

	noopHeader := http.Header{}
	SetTraceHeaders(opentracing.NoopTracer{}.StartSpan("noop"), noopHeader)
	assert.Empty(t, noopHeader)
}

func Test_TraceID(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tracer := newBridgeTracer(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	span := tracer.StartSpan("span")
	traceID := TraceID(span)
	span.Finish()
	assert.Len(t, traceID, 32)
	assert.Equal(t, recorder.Ended()[0].SpanContext().TraceID().String(), traceID)
	assert.Empty(t, TraceID(opentracing.NoopTracer{}.StartSpan("noop")))
}

func Test_Flush(t *testing.T) {
	calls := 0
	SetFlush(func() { calls++ })
	Flush()
	Flush()
	assert.Equal(t, 1, calls)

	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	SetFlush(func() {
		close(started)
		<-release
	})
	go Flush()
	<-started
	go func() {
		Flush()
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("second Flush returned before the first one finished")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	<-done
}
