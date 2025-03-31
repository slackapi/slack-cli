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

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Hash is a mapped value for a specific input
type Hash string

// NewHash creates a hash from provided bytes
//
// Before creating a new hash, provided JSON inputs are standardized with sorted
// attributes so the same input values are encoded to a matching hash.
//
// Reference: https://stackoverflow.com/questions/18668652/how-to-produce-json-with-sorted-keys-in-go/61887446#61887446
func NewHash(bytes []byte) (h Hash) {
	var ifce any
	var err error
	defer func() {
		if err != nil {
			h = encode(bytes)
		}
	}()
	err = json.Unmarshal(bytes, &ifce)
	if err != nil {
		return
	}
	hash, err := json.Marshal(ifce)
	if err != nil {
		return
	}
	return encode(hash)
}

// Equals returns true if the hash is equal
func (h Hash) Equals(is Hash) bool {
	return h == is
}

// encode turns bytes into a unique hash
func encode(bytes []byte) Hash {
	sha := sha256.Sum256(bytes)
	encoding := hex.EncodeToString(sha[:])
	return Hash(encoding)
}
