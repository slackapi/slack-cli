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
	"io"
	"log"
	"time"

	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
)

// SetupTracer sets up the tracer to send data to honeycomb
func SetupTracer(isDev bool) (io.Closer, opentracing.Tracer) {
	collectorEndpoint := "https://slackb.com/traces/v1/jaeger"
	if isDev {
		collectorEndpoint = "https://dev.slackb.com/traces/v1/jaeger"
	}
	// Recommended configuration for production.
	jCfg := jaegercfg.Configuration{
		ServiceName: "slack-cli", // Don't change this.  Required to distinguish logs & traces coming from the CLI
		Disabled:    false,       // Keep tracer active so span contexts and trace IDs are still generated
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	// Create HTTP transport with a short timeout to prevent hangs on unreachable hosts
	sender := transport.NewHTTPTransport(collectorEndpoint, transport.HTTPTimeout(1*time.Second))
	reporter := jaeger.NewRemoteReporter(sender,
		jaeger.ReporterOptions.BufferFlushInterval(100*time.Millisecond),
		// Having a larger value here results in longer execution of every single CLI command
		// We may need to adjust the size here if we observe lost data in metrics.
		jaeger.ReporterOptions.QueueSize(1),
	)

	// Initialize tracer with a logger and a metrics factory
	tracer, jaegerCloser, traceErr := jCfg.NewTracer(jaegercfg.Reporter(reporter))
	if traceErr != nil {
		log.Fatalf("Could not initialize jaeger tracer: %s", traceErr.Error())
	}

	opentracing.SetGlobalTracer(tracer)
	return jaegerCloser, tracer
}
