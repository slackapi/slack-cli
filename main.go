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

package main

import (
	"context"
	"os"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/cmd"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/contextutil"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/ioutils"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/tracer"
	"github.com/uber/jaeger-client-go"
)

func main() {
	defer recoveryFunc()

	var ctx = context.Background()
	// TODO - Could we refactor this to cmd/root.go to initialize open tracing after the CLI flags are parsed?
	//      - This would allow us to choose the correct API host based on flags
	//      - Uncomment `isDevTarget` if we refactor to cmd/root.go and update to call `ResolveApiHost`
	// var isDevTarget = shared.NewClientFactory().AuthClient().UserDefaultAuthIsProd(ctx) // TODO - hack, remove shared.clients
	var jaegerCloser, tracer = tracer.SetupTracer(false) // Always setup open tracing on prod
	ctx = contextutil.AddTracerToContext(ctx, tracer)
	defer jaegerCloser.Close()

	// set various metrics we will be tracking on the context
	ctx = config.SetContextSessionID(ctx, uuid.New().String())
	// add slack-cli version to context
	ctx = contextutil.AddVersionToContext(ctx, version.Get())

	osStr := os.Args[0:]
	processName := goutils.RedactPII(strings.Join(osStr, " "))

	var span = tracer.StartSpan("main", opentracing.Tag{Key: "version", Value: version.Get()})
	span.SetTag("slack_cli_sessionID", config.GetContextSessionID(ctx))
	span.SetTag("hashed_hostname", ioutils.GetHostname())
	span.SetTag("slack_cli_process", processName)
	// system_id is set in root.go initConfig()
	// project_id is set in root.go initConfig()

	if jaegerSpanContext, ok := span.Context().(jaeger.SpanContext); ok {
		ctx = config.SetContextTraceID(ctx, jaegerSpanContext.TraceID().String())
	}

	defer span.Finish()

	// add root span to context
	ctx = opentracing.ContextWithSpan(ctx, span)

	rootCmd, clients := cmd.Init()
	cmd.Execute(ctx, rootCmd, clients)
}

func recoveryFunc() {
	// in the event of a panic, log panic
	if r := recover(); r != nil {
		var clients = shared.NewClientFactory(shared.SetVersion(version.Raw()))
		var ctx = context.Background()
		ctx = config.SetContextSessionID(ctx, uuid.New().String())
		// set host for logging
		clients.Config.LogstashHostResolved = clients.AuthInterface().ResolveLogstashHost(ctx, clients.Config.ApiHostResolved, clients.Config.Version)
		clients.IO.PrintError(ctx, "Recovered from panic: %s\n%s", r, string(debug.Stack()))
		os.Exit(int(iostreams.ExitError))
	}
}
