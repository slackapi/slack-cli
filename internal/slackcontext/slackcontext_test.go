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

package slackcontext

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/tracer"
	"github.com/stretchr/testify/require"
)

func Test_SlackContext_OpenTracingSpan(t *testing.T) {
	tests := map[string]struct {
		expectedSpan  opentracing.Span
		expectedError error
	}{
		"returns span when span exists": {
			expectedSpan:  opentracing.StartSpan("test.span"),
			expectedError: nil,
		},
		"returns error when span not found": {
			expectedSpan:  nil,
			expectedError: slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for OpenTracing Span could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = opentracing.ContextWithSpan(ctx, tt.expectedSpan)
			actualSpan, actualError := OpenTracingSpan(ctx)
			require.Equal(t, tt.expectedSpan, actualSpan)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetOpenTracingSpan(t *testing.T) {
	tests := map[string]struct {
		expectedSpan opentracing.Span
	}{
		"sets the span in the context": {
			expectedSpan: opentracing.StartSpan("test.span"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetOpenTracingSpan(ctx, tt.expectedSpan)
			actualSpan := opentracing.SpanFromContext(ctx)
			require.Equal(t, tt.expectedSpan, actualSpan)
		})
	}
}

func Test_SlackContext_OpenTracingTraceID(t *testing.T) {
	tests := map[string]struct {
		expectedTraceID string
		expectedError   error
	}{
		"returns Trace ID when it exists": {
			expectedTraceID: uuid.New().String(),
			expectedError:   nil,
		},
		"returns error when Trace ID not found": {
			expectedTraceID: "",
			expectedError:   slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for OpenTracing Trace ID could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeyOpenTracingTraceID, tt.expectedTraceID)
			actualTraceID, actualError := OpenTracingTraceID(ctx)
			require.Equal(t, tt.expectedTraceID, actualTraceID)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetOpenTracingTraceID(t *testing.T) {
	tests := map[string]struct {
		expectedTraceID string
	}{
		"sets the Trace ID in the context": {
			expectedTraceID: uuid.New().String(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetOpenTracingTraceID(ctx, tt.expectedTraceID)
			actualTraceID := ctx.Value(contextKeyOpenTracingTraceID).(string)
			require.Equal(t, tt.expectedTraceID, actualTraceID)
		})
	}
}

func Test_SlackContext_OpenTracingTracer(t *testing.T) {
	_, _tracer := tracer.SetupTracer(true)

	tests := map[string]struct {
		expectedTracer opentracing.Tracer
		expectedError  error
	}{
		"returns Tracer when it exists": {
			expectedTracer: _tracer,
			expectedError:  nil,
		},
		"returns error when Tracer not found": {
			expectedTracer: nil,
			expectedError:  slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for OpenTracing Tracer could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeyOpenTracingTracer, tt.expectedTracer)
			actualTracer, actualError := OpenTracingTracer(ctx)
			require.Equal(t, tt.expectedTracer, actualTracer)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetOpenTracingTracer(t *testing.T) {
	_, _tracer := tracer.SetupTracer(true)

	tests := map[string]struct {
		expectedTracer opentracing.Tracer
	}{
		"sets the Tracer in the context": {
			expectedTracer: _tracer,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetOpenTracingTracer(ctx, tt.expectedTracer)
			actualTracer := ctx.Value(contextKeyOpenTracingTracer).(opentracing.Tracer)
			require.Equal(t, tt.expectedTracer, actualTracer)
		})
	}
}

func Test_SlackContext_ProjectID(t *testing.T) {
	tests := map[string]struct {
		expectedProjectID string
		expectedError     error
	}{
		"returns Project ID when it exists": {
			expectedProjectID: uuid.New().String(),
			expectedError:     nil,
		},
		"returns error when Project ID not found": {
			expectedProjectID: "",
			expectedError:     slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for Project ID could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeyProjectID, tt.expectedProjectID)
			actualProjectID, actualError := ProjectID(ctx)
			require.Equal(t, tt.expectedProjectID, actualProjectID)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetProjectID(t *testing.T) {
	tests := map[string]struct {
		expectedProjectID string
	}{
		"sets the Project ID in the context": {
			expectedProjectID: uuid.New().String(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetProjectID(ctx, tt.expectedProjectID)
			actualProjectID := ctx.Value(contextKeyProjectID).(string)
			require.Equal(t, tt.expectedProjectID, actualProjectID)
		})
	}
}

func Test_SlackContext_SessionID(t *testing.T) {
	tests := map[string]struct {
		expectedSessionID string
		expectedError     error
	}{
		"returns Session ID when it exists": {
			expectedSessionID: uuid.New().String(),
			expectedError:     nil,
		},
		"returns error when Session ID not found": {
			expectedSessionID: "",
			expectedError:     slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for Session ID could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeySessionID, tt.expectedSessionID)
			actualSessionID, actualError := SessionID(ctx)
			require.Equal(t, tt.expectedSessionID, actualSessionID)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetSessionID(t *testing.T) {
	tests := map[string]struct {
		expectedSessionID string
	}{
		"sets the Session ID in the context": {
			expectedSessionID: uuid.New().String(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetSessionID(ctx, tt.expectedSessionID)
			actualSessionID := ctx.Value(contextKeySessionID).(string)
			require.Equal(t, tt.expectedSessionID, actualSessionID)
		})
	}
}

func Test_SlackContext_SystemID(t *testing.T) {
	tests := map[string]struct {
		expectedSystemID string
		expectedError    error
	}{
		"returns System ID when it exists": {
			expectedSystemID: uuid.New().String(),
			expectedError:    nil,
		},
		"returns error when System ID not found": {
			expectedSystemID: "",
			expectedError:    slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for System ID could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeySystemID, tt.expectedSystemID)
			actualSystemID, actualError := SystemID(ctx)
			require.Equal(t, tt.expectedSystemID, actualSystemID)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetSystemID(t *testing.T) {
	tests := map[string]struct {
		expectedSystemID string
	}{
		"sets the System ID in the context": {
			expectedSystemID: uuid.New().String(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetSystemID(ctx, tt.expectedSystemID)
			actualSystemID := ctx.Value(contextKeySystemID).(string)
			require.Equal(t, tt.expectedSystemID, actualSystemID)
		})
	}
}

func Test_SlackContext_Version(t *testing.T) {
	tests := map[string]struct {
		expectedVersion string
		expectedError   error
	}{
		"returns Version when it exists": {
			expectedVersion: "v1.2.3",
			expectedError:   nil,
		},
		"returns error when Version not found": {
			expectedVersion: "",
			expectedError:   slackerror.New(slackerror.ErrContextValueNotFound).WithMessage("The value for Version could not be found"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = context.WithValue(ctx, contextKeyVersion, tt.expectedVersion)
			actualVersion, actualError := Version(ctx)
			require.Equal(t, tt.expectedVersion, actualVersion)
			require.Equal(t, tt.expectedError, actualError)
		})
	}
}

func Test_SlackContext_SetVersion(t *testing.T) {
	tests := map[string]struct {
		expectedVersion string
	}{
		"sets the Version in the context": {
			expectedVersion: "v1.2.3",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			ctx = SetVersion(ctx, tt.expectedVersion)
			actualVersion := ctx.Value(contextKeyVersion).(string)
			require.Equal(t, tt.expectedVersion, actualVersion)
		})
	}
}
