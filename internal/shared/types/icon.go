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

package types

import "encoding/json"

type Icons map[string]string

// MarshalJSON ensures we only return icon sizes of 192 and 96 as these
// are the only sizes guaranteed to be available synchronously (through apps.create/apps.edit).
// All sizes are stored in our db and available eventually (within ~10s usually).
func (i Icons) MarshalJSON() ([]byte, error) {
	// always return non-omitted, non-null object
	return json.Marshal(map[string]string{
		"image_96":  i["image_96"],
		"image_192": i["image_192"],
	})
}

func (i *Icons) UnmarshalJSON(data []byte) error {
	var tmp map[string]string
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if tmp["image_96"] == "" && tmp["image_192"] == "" {
		return nil
	}

	if *i == nil {
		*i = make(Icons)
	}

	for key, value := range tmp {
		(*i)[key] = value
	}

	return nil
}
