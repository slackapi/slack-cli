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

package create

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type SamplerMock struct {
	mock.Mock
}

func NewMockSampler() *SamplerMock {
	return &SamplerMock{}
}

func (s *SamplerMock) Do(req *http.Request) (*http.Response, error) {
	args := s.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
