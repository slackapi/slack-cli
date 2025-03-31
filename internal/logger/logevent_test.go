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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DataToString(t *testing.T) {
	logEvent := &LogEvent{
		Data: LogData{},
	}

	// Test empty key
	require.Equal(t, logEvent.DataToString(""), "", "should return the default value")

	// Test accessing a string
	logEvent.Data["hello"] = "world"
	require.Equal(t, logEvent.DataToString("hello"), "world", "should return the value")
	require.Equal(t, logEvent.DataToString("not-a-key"), "", "should return default value")

	// Test accessing invalid data type
	logEvent.Data["number"] = 100
	require.Equal(t, logEvent.DataToString("number"), "", "should return default value")
}

func Test_DataToStringSlice(t *testing.T) {
	logEvent := &LogEvent{
		Data: LogData{},
	}

	// Test empty key
	require.Equal(t, logEvent.DataToStringSlice(""), []string{}, "should return the default value")

	// Test accessing a string value in slice
	logEvent.Data["hello"] = []string{"world", "goodbye"}
	require.Contains(t, logEvent.DataToStringSlice("hello"), "goodbye", "should contain the value")

	// Test invalid key
	require.Equal(t, logEvent.DataToStringSlice("not-a-key"), []string{}, "should return default value")

	// Test accessing invalid data type
	logEvent.Data["number"] = 100
	require.Equal(t, logEvent.DataToStringSlice("number"), []string{}, "should return default value")
}

func Test_DataToInt(t *testing.T) {
	logEvent := &LogEvent{
		Data: LogData{},
	}

	// Test empty key
	require.Equal(t, logEvent.DataToInt(""), 0, "should return the default value")

	// Test accessing an int
	logEvent.Data["numRings"] = 7
	require.Equal(t, logEvent.DataToInt("numRings"), 7, "should contain the value")
	require.Equal(t, logEvent.DataToInt("not-a-key"), 0, "should return default value")

	// Test accessing invalid data type
	logEvent.Data["string"] = "i exist"
	require.Equal(t, logEvent.DataToInt("string"), 0, "should return default value")
}

func Test_DataToBool(t *testing.T) {
	logEvent := &LogEvent{
		Data: LogData{},
	}

	// Test empty key
	require.Equal(t, logEvent.DataToBool(""), false, "should return the default value")

	// Test accessing a string
	logEvent.Data["hello"] = true
	require.Equal(t, logEvent.DataToBool("hello"), true, "should return the value")
	require.Equal(t, logEvent.DataToBool("not-a-key"), false, "should return default value")

	// Test accessing invalid data type
	logEvent.Data["number"] = 100
	require.Equal(t, logEvent.DataToBool("number"), false, "should return default value")
}
