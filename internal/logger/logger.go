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

// Helper function definitions
type eventFunc func(event *LogEvent)

// Logger allows you to log events that another package can subscribe to and act upon.
// TODO - Can Logger be private requiring `logger.New` to be used? How do we pass instances of `logger` when private?
type Logger struct {
	// Public metadata that is attached to each log event
	Data LogData

	// Private output function where the log events are sent
	onEvent eventFunc
}

// Success return an event with the name "success"
func (l *Logger) SuccessEvent() *LogEvent {
	return &LogEvent{
		Level: "info",
		Name:  "success",
		Data:  l.Data,
	}
}

// Log event with a log level and log event name
func (l *Logger) Log(level string, name string) {
	if l.onEvent != nil {
		l.onEvent(&LogEvent{
			Level: level,
			Name:  name,
			Data:  l.Data,
		})
	}
}

// Debug log helper method
func (l *Logger) Debug(code string) {
	l.Log("debug", code)
}

// Info log helper method
func (l *Logger) Info(code string) {
	l.Log("info", code)
}

// Warn log helper method
func (l *Logger) Warn(code string) {
	l.Log("warn", code)
}

// Logger constructor that binds to an error, success, and event handler
func New(onEvent eventFunc) *Logger {
	return &Logger{
		onEvent: onEvent,
		Data:    LogData{},
	}
}
