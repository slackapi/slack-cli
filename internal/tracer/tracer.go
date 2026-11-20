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
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// SetupTracer sets up the tracer to send data to honeycomb with OpenTelemetry
func SetupTracer(isDev bool) (io.Closer, opentracing.Tracer) {
	var collectorEndpoint = "slackb.com"
	if isDev {
		collectorEndpoint = "dev.slackb.com"
	}
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(collectorEndpoint),
		otlptracehttp.WithURLPath("/traces/v1/otlp"),
	)
	if err != nil {
		log.Fatalf("Could not initialize the trace exporter: %s", err.Error())
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		// No host detector: spans must not carry the raw hostname or IP
		sdktrace.WithResource(resource.NewSchemaless(
			// Don't change these. Required to distinguish logs & traces coming from the CLI
			attribute.String("service.name", "slack-cli"),
			// The slackb.com Jaeger endpoint set this tag, so Honeycomb queries use it
			attribute.String("service_name", "slack-cli"),
		)),
	)
	// Like the Jaeger reporter, don't print export errors from a blocked network
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(error) {}))
	bridge := newBridgeTracer(provider)
	opentracing.SetGlobalTracer(bridge)
	return closerFunc(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return provider.Shutdown(ctx)
	}), bridge
}

// newBridgeTracer returns an OpenTracing tracer that records spans with provider
func newBridgeTracer(provider trace.TracerProvider) *otelbridge.BridgeTracer {
	bridge := otelbridge.NewBridgeTracer()
	bridge.SetOpenTelemetryTracer(provider.Tracer("slack-cli"))
	bridge.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))
	otel.SetTracerProvider(otelbridge.NewTracerProvider(bridge, provider))
	return bridge
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// SetTraceHeaders adds the x-b3 headers for span to header. The sampled header
// is always "0" so the Slack API doesn't sample requests from the CLI.
func SetTraceHeaders(span opentracing.Span, header http.Header) {
	carrier := opentracing.HTTPHeadersCarrier(http.Header{})
	if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
		return
	}
	for key, values := range carrier {
		header[key] = values
	}
	if header.Get("x-b3-traceid") != "" {
		header.Set("x-b3-sampled", "0")
	}
}

// TraceID returns the trace ID of span, or "" if span has no trace ID.
func TraceID(span opentracing.Span) string {
	header := http.Header{}
	SetTraceHeaders(span, header)
	return header.Get("x-b3-traceid")
}

var (
	flushMu   sync.Mutex
	flushFunc func()
)

// SetFlush sets the function that Flush calls to finish and send spans.
func SetFlush(f func()) {
	flushMu.Lock()
	defer flushMu.Unlock()
	flushFunc = f
}

// Flush calls the function from SetFlush once and then forgets it, so every
// exit path can call Flush without sending spans twice. A concurrent caller
// waits for the first flush to finish before it returns.
func Flush() {
	flushMu.Lock()
	defer flushMu.Unlock()
	if flushFunc != nil {
		flushFunc()
		flushFunc = nil
	}
}
