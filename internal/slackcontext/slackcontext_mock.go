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

	"github.com/google/uuid"
)

const (
	mockCLIVersion = "v1.2.3"
)

// MockContext sets values in the context that are guaranteed to exist before
// the Cobra root command is executed.
func MockContext(ctx context.Context) context.Context {
	ctx = SetOpenTracingTraceID(ctx, uuid.New().String())
	ctx = SetSessionID(ctx, uuid.New().String())
	ctx = SetVersion(ctx, mockCLIVersion)
	return ctx
}
