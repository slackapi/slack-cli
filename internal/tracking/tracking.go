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

package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/ioutils"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/style"
)

// TrackingManager is an interface for tracking metrics and events related to CLI activity
type TrackingManager interface {
	FlushToLogstash(ctx context.Context, cfg *config.Config, io iostreams.IOStreamer, exitCode iostreams.ExitCode) error
	getSessionData() EventData
	cleanSessionData(EventData) EventData
	// Setter methods for metric fields
	SetErrorCode(code string)
	SetErrorMessage(err string)
	SetAuthEnterpriseID(id string)
	SetAuthTeamID(id string)
	SetAuthUserID(id string)
	SetAppEnterpriseID(id string)
	SetAppTeamID(id string)
	SetAppUserID(id string)
	SetAppTemplate(template string)
}

type EventTracker struct {
	mu        sync.RWMutex
	eventData EventData
}

// NewEventTracker returns an EventTracker instance, for tracking event-related data and metrics
func NewEventTracker() *EventTracker {
	eventTracker := &EventTracker{
		mu: sync.RWMutex{},
		eventData: EventData{
			App:  AppEventData{},
			Auth: AuthEventData{},
		},
	}
	return eventTracker
}

// getSessionData DO NOT USE THIS! It may be tempting but don't do it!
func (e *EventTracker) getSessionData() EventData {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.eventData
}

// setSessionData sets the session data object
func (e *EventTracker) setSessionData(data EventData) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eventData = data
}

// SetErrorCode sets the error code associated to this CLI execution for metrics
func (e *EventTracker) SetErrorCode(code string) {
	data := e.getSessionData()
	data.ErrorCode = code
	e.setSessionData(data)
}

// SetErrorMessage sets the error message associated to this CLI execution for metrics
func (e *EventTracker) SetErrorMessage(err string) {
	data := e.getSessionData()
	data.ErrorMessage = err
	e.setSessionData(data)
}

// SetAuthEnterpriseID sets the auth enterprise ID associated to this CLI execution for metrics
func (e *EventTracker) SetAuthEnterpriseID(id string) {
	data := e.getSessionData()
	data.Auth.EnterpriseID = id
	e.setSessionData(data)
}

// SetAuthTeamID sets the auth team ID associated to this CLI execution for metrics
func (e *EventTracker) SetAuthTeamID(id string) {
	data := e.getSessionData()
	data.Auth.TeamID = id
	e.setSessionData(data)
}

// SetAuthTeamID sets the auth team ID associated to this CLI execution for metrics
func (e *EventTracker) SetAuthUserID(id string) {
	data := e.getSessionData()
	data.Auth.UserID = id
	e.setSessionData(data)
}

// SetAppEnterpriseID sets the app enterprise ID associated to this CLI execution for metrics
func (e *EventTracker) SetAppEnterpriseID(id string) {
	data := e.getSessionData()
	data.App.EnterpriseID = id
	e.setSessionData(data)
}

// SetAppTeamID sets the app team ID associated to this CLI execution for metrics
func (e *EventTracker) SetAppTeamID(id string) {
	data := e.getSessionData()
	data.App.TeamID = id
	e.setSessionData(data)
}

// SetAppUserID sets the app user ID associated to this CLI execution for metrics
func (e *EventTracker) SetAppUserID(id string) {
	data := e.getSessionData()
	data.App.UserID = id
	e.setSessionData(data)
}

// SetAppTemplate sets the app template used in this CLI execution for metrics
func (e *EventTracker) SetAppTemplate(template string) {
	data := e.getSessionData()
	data.App.Template = template
	e.setSessionData(data)
}

// cleanSessionData ensures every string value the provided object has PII redacted
func (e *EventTracker) cleanSessionData(data EventData) EventData {
	if len(data.ErrorMessage) > 0 {
		data.ErrorMessage = goutils.RedactPII(style.RemoveANSI(data.ErrorMessage))
	}
	if len(data.App.Template) > 0 {
		data.App.Template = goutils.RedactPII(data.App.Template)
	}

	return data
}

// FlushToLogstash will send an event representing this session to logstash
func (e *EventTracker) FlushToLogstash(ctx context.Context, cfg *config.Config, ioStream iostreams.IOStreamer, exitCode iostreams.ExitCode) error {
	if cfg.DisableTelemetryFlag {
		return nil
	}
	postURL := cfg.LogstashHostResolved
	if postURL == "" {
		// Root command initialization was not run; we might get here if user ran `slack --version`.
		// In this case, the root command was not initialized, so none of the bootup routine executed (flags weren't parsed, config not initialized, etc.).
		return nil
	}
	versionString, _ := strings.CutPrefix(cfg.Version, "v")
	eventData := e.cleanSessionData(e.getSessionData())
	sessionID, err := slackcontext.SessionID(ctx)
	if err != nil {
		return err
	}

	var eventName EventType
	switch exitCode {
	case iostreams.ExitCancel:
		eventName = Interrupt
	case iostreams.ExitError:
		eventName = Error
	default:
		eventName = Success
	}

	var event = LogstashEvent{
		Event:     eventName,
		Timestamp: time.Now().UnixMilli(),
		Data:      eventData,
		Context: EventContext{
			CLIVersion:       versionString,
			Host:             ioutils.GetHostname(),
			OS:               runtime.GOOS,
			SessionID:        sessionID,
			SystemID:         cfg.SystemID,
			ProjectID:        cfg.ProjectID,
			Binary:           goutils.RedactPII(strings.Join(os.Args[0:1], "")),
			Command:          cfg.Command,
			CommandCanonical: cfg.CommandCanonical,
			Flags:            cfg.RawFlags,
			Runtime:          cfg.RuntimeName,
			RuntimeVersion:   cfg.RuntimeVersion,
		},
	}

	postBody, err := json.Marshal([]LogstashEvent{event})
	if err != nil {
		return err
	}

	ioStream.PrintDebug(ctx, "FlushToLogstash will POST %s payload: %s", postURL, string(postBody))
	responseBody := bytes.NewBuffer(postBody)

	request, err := http.NewRequest("POST", postURL, responseBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("X-Slack-Ses-Id", sessionID)

	client := &http.Client{Timeout: time.Second * 1}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	ioStream.PrintDebug(ctx, "FlushToLogstash response status code: %d, body: %s", response.StatusCode, string(b))

	return nil
}
