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

package main

import (
	"context"
	"os"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/slackapi/slack-cli/cmd"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/ioutils"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/tracer"
	"github.com/slackapi/slack-cli/internal/useragent"
	"github.com/slackapi/slack-cli/internal/version"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	defer recoveryFunc()

	// Create the parent context for the CLI execution
	var ctx = context.Background()

	// TODO - Could we refactor this to cmd/root.go to initialize open tracing after the CLI flags are parsed?
	//      - This would allow us to choose the correct API host based on flags
	//      - Uncomment `isDevTarget` if we refactor to cmd/root.go and update to call `ResolveAPIHost`
	// var isDevTarget = shared.NewClientFactory().AuthClient().UserDefaultAuthIsProd(ctx) // TODO - hack, remove shared.clients
	traceCloser := tracer.SetupTracer(false) // Always setup tracing on prod

	// Set context values
	sessionID := uuid.New().String()
	cliVersion := version.Raw()
	ctx = slackcontext.SetSessionID(ctx, sessionID)
	ctx = slackcontext.SetVersion(ctx, cliVersion)

	osStr := os.Args[0:]
	processName := goutils.RedactPII(strings.Join(osStr, " "))

	ctx, span := tracer.Start(ctx, "main")
	span.SetAttributes(
		attribute.String("version", cliVersion),
		attribute.String("slack_cli_sessionID", sessionID),
		attribute.String("hashed_hostname", ioutils.GetHostname()),
		attribute.String("slack_cli_process", processName),
	)
	if aiAgentName := useragent.GetAIAgentName(); aiAgentName != "" {
		span.SetAttributes(attribute.String("ai_agent", aiAgentName))
	}
	// system_id is set in root.go initConfig()
	// project_id is set in root.go initConfig()

	if spanContext := span.SpanContext(); spanContext.IsValid() {
		ctx = slackcontext.SetTraceID(ctx, spanContext.TraceID().String())
	}

	// Flush on return here, and before each exit in cmd/root.go, which skips these defers
	tracer.SetFlush(func() {
		span.End()
		_ = traceCloser.Close()
	})
	defer tracer.Flush()

	rootCmd, clients := cmd.Init(ctx)
	cmd.ExecuteContext(ctx, rootCmd, clients)
}

// TODO(slackcontext) Use closure to pass in the ctx, which includes the sessionID
func recoveryFunc() {
	// in the event of a panic, log panic
	if r := recover(); r != nil {
		var clients = shared.NewClientFactory()

		var ctx = context.Background()
		ctx = slackcontext.SetSessionID(ctx, uuid.New().String())
		ctx = slackcontext.SetVersion(ctx, version.Raw())

		// set host for logging
		clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved)
		clients.IO.PrintError(ctx, "Recovered from panic: %s\n%s", r, string(debug.Stack()))
		os.Exit(int(iostreams.ExitError))
	}
}
