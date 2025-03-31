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

package auth

import (
	"context"
)

// RevokeToken revokes API access of the given token
//
// An error is returned if the revoking request failed for a reason unrelated to
// expired or a somehow invalid authentication. That error should not be ignored
// since it might mean the token was not revoked.
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	err := c.api.RevokeToken(ctx, token)
	if err != nil {
		_, unfilteredError := c.FilterKnownAuthErrors(ctx, err)
		if unfilteredError != nil {
			return unfilteredError
		}
	}
	return nil
}
