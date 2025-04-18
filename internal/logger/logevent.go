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

package logger

// LogEvent is created for each log that occurs and explains details about the log message.
type LogEvent struct {
	// Log levels:
	// - debug
	// - info
	// - warn
	// - error
	Level string

	// Identifies the type of event with a unique name
	Name string

	// Metadata available to the log event
	Data LogData
}

// DataToString safely returns the Data[key] value as a string
func (l *LogEvent) DataToString(key string) string {
	var defaultValue string

	// Check that data exists
	v := l.Data[key]
	if v == nil {
		return defaultValue
	}

	// Attempt casting
	if s, ok := v.(string); ok {
		return s
	} else {
		return defaultValue
	}
}

// DataToString safely returns the Data[key] value as a string
func (l *LogEvent) DataToStringSlice(key string) []string {
	defaultValue := make([]string, 0)

	// Check that data exists
	v := l.Data[key]
	if v == nil {
		return defaultValue
	}

	// Attempt casting
	if s, ok := v.([]string); ok {
		return s
	} else {
		return defaultValue
	}
}

// DataToBool safely returns the Data[key] value as a bool
func (l *LogEvent) DataToBool(key string) bool {
	var defaultValue bool

	// Check that data exists
	v := l.Data[key]
	if v == nil {
		return defaultValue
	}

	// Attempt casting
	if s, ok := v.(bool); ok {
		return s
	} else {
		return defaultValue
	}
}

// DataToBool safely returns the Data[key] value as a bool
func (l *LogEvent) DataToInt(key string) int {
	var defaultValue int

	// Check that data exists
	v := l.Data[key]
	if v == nil {
		return defaultValue
	}

	// Attempt casting
	if s, ok := v.(int); ok {
		return s
	} else {
		return defaultValue
	}
}
