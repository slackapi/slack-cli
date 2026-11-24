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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

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

func Test_newProvider(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "authorization=secret")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS", "authorization=secret")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "host.name=leaked-hostname")
	t.Setenv("OTEL_TRACES_SAMPLER", "always_off")
	t.Setenv("OTEL_BSP_MAX_QUEUE_SIZE", "0")
	type request struct {
		header http.Header
		body   []byte
	}
	received := make(chan request, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received <- request{r.Header, body}
	}))
	defer server.Close()
	options := append(exporterOptions(server.URL+"/traces/v1/otlp"),
		otlptracehttp.WithTLSClientConfig(server.Client().Transport.(*http.Transport).TLSClientConfig))
	exporter, err := otlptracehttp.New(context.Background(), options...)
	require.NoError(t, err)
	provider := newProvider(exporter)
	_, span := provider.Tracer("test").Start(context.Background(), "span")
	span.End()
	require.NoError(t, provider.ForceFlush(context.Background()))
	select {
	case r := <-received:
		assert.Empty(t, r.header.Get("authorization"))
		assert.Contains(t, string(r.body), "service_name")
		assert.NotContains(t, string(r.body), "leaked-hostname")
	case <-time.After(5 * time.Second):
		t.Fatal("no export reached the TLS server")
	}
}

func Test_newTracerProvider(t *testing.T) {
	tests := map[string]struct {
		env           map[string]string
		expectedCalls int
	}{
		"builds the exporter when telemetry is on":        {env: map[string]string{}, expectedCalls: 1},
		"never builds the exporter when telemetry is off": {env: map[string]string{"SLACK_DISABLE_TELEMETRY": "true"}, expectedCalls: 0},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			exporter := tracetest.NewInMemoryExporter()
			provider, err := newTracerProvider(func(key string) string { return tc.env[key] }, func() (sdktrace.SpanExporter, error) {
				calls++
				return exporter, nil
			})
			require.NoError(t, err)
			_, span := provider.Tracer("test").Start(context.Background(), "span")
			span.End()
			require.NoError(t, provider.ForceFlush(context.Background()))
			assert.Equal(t, tc.expectedCalls, calls)
			assert.Len(t, exporter.GetSpans(), tc.expectedCalls)
			assert.True(t, span.SpanContext().IsValid())
		})
	}
}

func Test_SetTraceHeaders(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	ctx, parent := provider.Tracer("test").Start(context.Background(), "parent")
	ctx, span := provider.Tracer("test").Start(ctx, "child")
	header := http.Header{}
	SetTraceHeaders(ctx, header)
	span.End()
	parent.End()
	assert.Equal(t, "0", header.Get("x-b3-sampled"))
	assert.Equal(t, span.SpanContext().SpanID().String(), header.Get("x-b3-spanid"))
	assert.Equal(t, span.SpanContext().TraceID().String(), header.Get("x-b3-traceid"))

	noopHeader := http.Header{}
	SetTraceHeaders(context.Background(), noopHeader)
	assert.Empty(t, noopHeader)
}

func Test_Start(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tracer := newBridgeTracer(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	opentracing.SetGlobalTracer(tracer)
	defer opentracing.SetGlobalTracer(opentracing.NoopTracer{})
	parent := tracer.StartSpan("opentracing-parent")
	ctx := opentracing.ContextWithSpan(context.Background(), parent)
	ctx, child := Start(ctx, "otel-child")
	grandchild, _ := opentracing.StartSpanFromContext(ctx, "opentracing-grandchild")
	grandchild.Finish()
	child.End()
	parent.Finish()
	spans := map[string]sdktrace.ReadOnlySpan{}
	for _, span := range recorder.Ended() {
		spans[span.Name()] = span
	}
	require.Len(t, spans, 3)
	assert.Equal(t, spans["opentracing-parent"].SpanContext().SpanID(), spans["otel-child"].Parent().SpanID())
	assert.Equal(t, spans["otel-child"].SpanContext().SpanID(), spans["opentracing-grandchild"].Parent().SpanID())
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
