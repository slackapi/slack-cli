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
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// SetupTracer sets up the tracer to send data to honeycomb with OpenTelemetry
func SetupTracer(isDev bool) io.Closer {
	var collectorEndpoint = "https://slackb.com/traces/v1/otlp"
	if isDev {
		collectorEndpoint = "https://dev.slackb.com/traces/v1/otlp"
	}
	// Like the Jaeger reporter, print nothing: not export errors from a blocked
	// network, and not diagnostics about the developer's own OTEL_* variables
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(error) {}))
	otel.SetLogger(logr.Discard())
	exporter, err := otlptracehttp.New(context.Background(), exporterOptions(collectorEndpoint)...)
	if err != nil {
		log.Fatalf("Could not initialize the trace exporter: %s", err.Error())
	}
	provider := newProvider(exporter)
	bridge := newBridgeTracer(provider)
	opentracing.SetGlobalTracer(bridge)
	return closerFunc(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return provider.Shutdown(ctx)
	})
}

// exporterOptions sets every value that OTEL_EXPORTER_OTLP_* environment
// variables can change, so a developer's own collector headers or settings are
// never sent to Slack
func exporterOptions(endpointURL string) []otlptracehttp.Option {
	return []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpointURL),
		otlptracehttp.WithHeaders(map[string]string{}),
		otlptracehttp.WithTLSClientConfig(&tls.Config{}),
		otlptracehttp.WithCompression(otlptracehttp.NoCompression),
		otlptracehttp.WithTimeout(10 * time.Second),
	}
}

// newProvider returns a tracer provider that batches spans to exporter. It
// ignores OTEL_RESOURCE_ATTRIBUTES, OTEL_TRACES_SAMPLER, OTEL_BSP_*, and OTEL_*_LIMIT from the environment.
func newProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// No host detector: spans must not carry the raw hostname or IP
	res := resource.NewSchemaless(
		// Don't change these. Required to distinguish logs & traces coming from the CLI
		attribute.String("service.name", "slack-cli"),
		// The slackb.com Jaeger endpoint set this tag, so Honeycomb queries use it
		attribute.String("service_name", "slack-cli"),
	)
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(resourceExporter{exporter, res},
			// The defaults, set here so OTEL_BSP_* environment variables can't change them
			sdktrace.WithMaxQueueSize(sdktrace.DefaultMaxQueueSize),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			sdktrace.WithExportTimeout(sdktrace.DefaultExportTimeout*time.Millisecond),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// The defaults, set here so OTEL_*_LIMIT environment variables can't change them
		sdktrace.WithRawSpanLimits(sdktrace.SpanLimits{
			AttributeValueLengthLimit:   sdktrace.DefaultAttributeValueLengthLimit,
			AttributeCountLimit:         sdktrace.DefaultAttributeCountLimit,
			EventCountLimit:             sdktrace.DefaultEventCountLimit,
			LinkCountLimit:              sdktrace.DefaultLinkCountLimit,
			AttributePerEventCountLimit: sdktrace.DefaultAttributePerEventCountLimit,
			AttributePerLinkCountLimit:  sdktrace.DefaultAttributePerLinkCountLimit,
		}),
	)
}

// resourceExporter exports spans with a fixed resource, because the provider's
// own resource always merges in OTEL_RESOURCE_ATTRIBUTES
type resourceExporter struct {
	sdktrace.SpanExporter
	resource *resource.Resource
}

func (e resourceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	fixed := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		fixed[i] = resourceSpan{span, e.resource}
	}
	return e.SpanExporter.ExportSpans(ctx, fixed)
}

type resourceSpan struct {
	sdktrace.ReadOnlySpan
	resource *resource.Resource
}

func (s resourceSpan) Resource() *resource.Resource { return s.resource }

// newBridgeTracer returns an OpenTracing tracer that records spans with provider
func newBridgeTracer(provider trace.TracerProvider) *otelbridge.BridgeTracer {
	bridge := otelbridge.NewBridgeTracer()
	// A wrapper tracer lets OpenTelemetry spans see OpenTracing parents in a context
	bridge.SetOpenTelemetryTracer(otelbridge.NewWrapperTracer(bridge, provider.Tracer("slack-cli")))
	bridge.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))
	otel.SetTracerProvider(otelbridge.NewTracerProvider(bridge, provider))
	return bridge
}

// Start starts an OpenTelemetry span named name as a child of the span in ctx,
// and returns a context that holds the new span. End the span when done.
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("slack-cli").Start(ctx, name)
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// The sampled header is always "0" so the Slack API doesn't sample requests from the CLI.
func SetTraceHeaders(ctx context.Context, header http.Header) {
	if !trace.SpanContextFromContext(ctx).IsValid() {
		return
	}
	b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)).Inject(ctx, propagation.HeaderCarrier(header))
	header.Set("x-b3-sampled", "0")
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
