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

package iostreams

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/style"
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
	span, _ := opentracing.StartSpanFromContext(ctx, "printDebug", opentracing.Tag{Key: "debug_log", Value: message})
	defer span.Finish()
	lines := strings.Split(message, "\n")
	for _, line := range lines {
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
	span, _ := opentracing.StartSpanFromContext(ctx, "printError", opentracing.Tag{Key: "error_log", Value: message})
	defer span.Finish()
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
	span, _ := opentracing.StartSpanFromContext(ctx, "printWarning", opentracing.Tag{Key: "warning_log", Value: message})
	defer span.Finish()
	_ = io.FlushToLogFile(ctx, "warning", message)
	io.Stderr.Println("\n" + style.Emoji("warning") + message)
}

// PrintInfo print a formatted message to stdout, sometimes tracing context
func (io *IOStreams) PrintInfo(ctx context.Context, shouldTrace bool, format string, a ...any) {
	message := sprintF(format, a...)
	if shouldTrace {
		span, _ := opentracing.StartSpanFromContext(ctx, "printInfo", opentracing.Tag{Key: "printInfo", Value: message})
		defer span.Finish()
	}
	io.Stdout.Println(style.Styler().Reset(message))
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
