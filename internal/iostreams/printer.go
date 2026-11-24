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

package iostreams

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/internal/tracer"
	"go.opentelemetry.io/otel/attribute"
)

// Printer contains implementations of printers that output to the display or
// file
//
// These are preferred methods of capturing printed text for mocking and testing
type Printer interface {
	// PrintDebug prints a debug message to stdout if debug is enabled
	PrintDebug(ctx context.Context, format string, a ...any)
	// PrintError logs and prints an error message to stderr
	PrintError(ctx context.Context, format string, a ...any)
	// PrintWarning logs and prints a warning message to stdout
	PrintWarning(ctx context.Context, format string, a ...any)
	// PrintInfo print a formatted message to stdout, sometimes tracing context
	PrintInfo(ctx context.Context, shouldTrace bool, format string, a ...any)
	// PrintTrace prints traceID and values to stdout if SLACK_TEST_TRACE=true
	//
	// Trace value definitions are listed in internal/slacktrace/slacktrace.go
	PrintTrace(ctx context.Context, traceID string, traceValues ...string)
}

// PrintDebug prints a debug message to stdout if debug is enabled
func (io *IOStreams) PrintDebug(ctx context.Context, format string, a ...any) {
	message := strings.TrimSpace(style.RemoveANSI(sprintF(format, a...)))
	_, span := tracer.Start(ctx, "printDebug")
	span.SetAttributes(attribute.String("debug_log", message))
	defer span.End()
	lines := strings.SplitSeq(message, "\n")
	for line := range lines {
		_ = io.FlushToLogFile(ctx, "debug", line)
		io.FinishLogFile(ctx)
		if io.config.DebugEnabled {
			debug := "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + line
			io.Stdout.Println(style.Secondary(debug))
		}
	}
}

// PrintError logs and prints an error message to stderr
func (io *IOStreams) PrintError(ctx context.Context, format string, a ...any) {
	message := sprintF(format, a...)
	_, span := tracer.Start(ctx, "printError")
	span.SetAttributes(attribute.String("error_log", message))
	defer span.End()
	_ = io.FlushToLogFile(ctx, "error", message)
	io.FinishLogFile(ctx)
	if !strings.Contains(message, style.Emoji("prohibited")) {
		io.Stderr.Println("\n" + style.Emoji("prohibited") + message)
	} else {
		io.Stderr.Println(message)
	}
}

// PrintWarning logs and prints a warning message to stdout
func (io *IOStreams) PrintWarning(ctx context.Context, format string, a ...any) {
	message := sprintF(format, a...)
	_, span := tracer.Start(ctx, "printWarning")
	span.SetAttributes(attribute.String("warning_log", message))
	defer span.End()
	_ = io.FlushToLogFile(ctx, "warning", message)
	io.Stderr.Println("\n" + style.Emoji("warning") + message)
}

// PrintInfo print a formatted message to stdout, sometimes tracing context
func (io *IOStreams) PrintInfo(ctx context.Context, shouldTrace bool, format string, a ...any) {
	message := sprintF(format, a...)
	if shouldTrace {
		_, span := tracer.Start(ctx, "printInfo")
		span.SetAttributes(attribute.String("printInfo", message))
		defer span.End()
	}
	if style.IsLipglossEnabled() {
		io.Stdout.Println(message)
	} else {
		io.Stdout.Println(style.Styler().Reset(message))
	}
}

// PrintTrace prints traceID and values to stdout if SLACK_TEST_TRACE=true
func (io *IOStreams) PrintTrace(ctx context.Context, traceID string, traceValues ...string) {
	if !io.config.SlackTestTraceFlag {
		return
	}
	io.Stdout.Println(style.Tracef(traceID, traceValues...))
}

// sprintF formats format with the existing args or returns an unchanged string
func sprintF(format string, args ...any) string {
	if len(args) != 0 {
		return fmt.Sprintf(format, args...)
	} else {
		return format
	}
}
