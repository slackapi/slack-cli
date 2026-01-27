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

package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/radovskyb/watcher"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// for lazy testing
var websocketDialerDial = func(d *websocket.Dialer, urlStr string,
	requestHeader http.Header) (WebSocketConnection, *http.Response, error) {
	return d.Dial(urlStr, requestHeader)
}

// WebSocketConnection interface representing interacting with a WebSocket connection - mockable
type WebSocketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	Close() error
}

// sendWebSocketCloseControlMessage Send a websocket close message to signal the start of closing the connection
func sendWebSocketCloseControlMessage(ctx context.Context, clients *shared.ClientFactory, conn WebSocketConnection) {
	if conn != nil {
		clients.IO.PrintDebug(ctx, "Sending WebSocket Close control message")
		if closeErr := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second)); closeErr != nil {
			clients.IO.PrintDebug(ctx, "Error sending websocket close control message: %s", closeErr.Error())
		}
	}
}

// LocalHostedContext describes the locally installed app and workspace
type LocalHostedContext struct {
	BotAccessToken string            `json:"bot_access_token,omitempty"`
	AppID          string            `json:"app_id,omitempty"`
	TeamID         string            `json:"team_id,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
}

// LocalServer runs the Slack app locally. It will try to Start a connection
// ahead of executing a script hook and returning responses. If sdk has indicated
// that it will manage the connection to Slack, LocalServer will delegate connection
// response management to the script hook
type LocalServer struct {
	clients            *shared.ClientFactory
	log                *logger.Logger
	token              string
	localHostedContext LocalHostedContext
	cliConfig          hooks.SDKCLIConfig
	Connection         WebSocketConnection
	delegateCmd        hooks.ShellCommand // track running delegated process
	delegateCmdMutex   sync.Mutex         // protect concurrent access
}

// Start establishes a socket connection to Slack, which will receive app-relevant events. It does so in a loop to support for re-establishing the socket connection.
func (r *LocalServer) Start(ctx context.Context) error {
	for {
		// Wrapping in an error function so that we can `defer` closing the TCP connection within the loop in the case of a restart
		err := func() error {
			// Get a socket connection address
			r.clients.IO.PrintDebug(ctx, "Retrieving and establishing connection to WebSocket URL...")
			result, err := r.clients.API().ConnectionsOpen(ctx, r.token)
			if err != nil {
				return slackerror.Wrap(err, slackerror.ErrSocketConnection).WithMessage("Error fetching socket connection URL")
			}

			// Open the websocket connection
			c, _, err := websocketDialerDial(websocket.DefaultDialer, result.URL, nil)
			if err != nil {
				return slackerror.Wrap(err, slackerror.ErrSocketConnection).WithMessage("Error establishing socket connection")
			}
			r.Connection = c
			// Signal to CLI that this command will need to do additional cleanup of I/O (closing socket connection cleanly); matching Done() in defer function below
			r.clients.CleanupWaitGroup.Add(1)
			// Two channels to communicate with Listen(): errChan for errors, and done for signaling restarting the connection
			// A special "clean exit" error exists for signaling a graceful exit; run.go handles this special error
			errChan := make(chan error)
			done := make(chan bool)
			go r.Listen(ctx, errChan, done)
			// Cleanup routine: close TCP connection and notify global waitgroup that we are done.
			defer func() {
				close(errChan)
				close(done)
				r.clients.IO.PrintDebug(ctx, "LocalServer.Start closing websocket TCP connection")
				r.Connection.Close()
				r.clients.CleanupWaitGroup.Done()
			}()
			// Wait for either an error (via errChan), or for Listen to finish cleanly (via done)
			select {
			case err := <-errChan:
				// If this is a clean exit, raise the special error code up
				if slackerror.Is(err, slackerror.ErrLocalAppRunCleanExit) {
					return err
				}
				r.clients.IO.PrintDebug(ctx, "LocalServer.Listen errored: %s", err.Error())
				sendWebSocketCloseControlMessage(ctx, r.clients, r.Connection)
				return slackerror.Wrap(err, slackerror.ErrLocalAppRun)
			case <-done:
				r.clients.IO.PrintDebug(ctx, "LocalServer.Listen signalled for restart")
				return nil
			}
		}()
		if err != nil {
			return err
		}
	}
}

// Listen waits for incoming events over a socket connection and invokes the deno-sdk-powered app with each payload. Responds to each event in a way that mimics the behaviour of a hosted app.
func (r *LocalServer) Listen(ctx context.Context, errChan chan<- error, done chan<- bool) {
	r.log.Info("on_cloud_run_connection_connected")

	// Listen for socket messages
	for {
		select {
		case <-ctx.Done():
			// In case execution _happens_ to be here before calling ReadMessage() below, and the user ctrl+c,
			// we can exit early and cleanly. Very unlikely, though, as ReadMessage() below blocks.
			errChan <- slackerror.New(slackerror.ErrLocalAppRunCleanExit)
			return
		default:
			// Unfortunately, the following call blocks the thread
			_, messageBytes, err := r.Connection.ReadMessage()
			if err != nil {
				// If Slack backend signals that it is going down (CloseGoingAway), or the connection is being terminated by Slack - possibly as a result of user ctrl+c and us sending a Close Control Message and Slack echoing it back to us as part of a goodbye handshake (CloseNormalClosure),
				// gorilla will by default send a close message back as per the websocket spec. In this case, we only need to close the TCP connection,
				// which is done via `defer` in LocalServer.Start()
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					r.clients.IO.PrintDebug(ctx, "Received a WebSocket Close Control message : %s; will close connection shortly", err.Error())
					// Signal the error channel for a clean exit.
					errChan <- slackerror.New(slackerror.ErrLocalAppRunCleanExit)
				} else {
					errChan <- slackerror.Wrap(err, slackerror.ErrSocketConnection)
				}
				return
			}
			r.log.Data["cloud_run_connection_message"] = string(messageBytes)
			r.log.Debug("on_cloud_run_connection_message")

			var msg Message
			err = json.Unmarshal(messageBytes, &msg)
			if err != nil {
				// If we get here then we received an unexpected response payload from the server
				// we do not want to error out but instead, re-start the connection.
				// Choosing this route because sometimes the server returns with errors like `UNAUTHENTICATED: cache_error`
				// in that case, we do not want to exit the whole local run experience but merely warn the user and re-connect
				r.clients.IO.PrintDebug(ctx, "Re-establishing socket connection as we received an unexpected response from server: %s", string(messageBytes))
				done <- false
				return
			}

			var linkResponse *LinkResponse
			switch msg.Type {
			case helloMessageType:
				// ignore any hello messages from the server
				continue
			case disconnectMessageType:
				// when we receive a disconnect event, we should reconnect
				done <- false
				return
			default:
				var socketEvent = SocketEvent{
					Body:    msg.Payload,
					Context: r.localHostedContext,
				}

				body, err := json.Marshal(socketEvent)
				if err != nil {
					errChan <- slackerror.Wrap(err, slackerror.ErrSocketConnection)
					return
				}

				_, err = r.clients.SDKConfig.Hooks.Start.Get()
				if err != nil {
					errChan <- err
					return
				}
				// Mimic the hosted app by executing the SDKs run command with the message as a param
				var startHookOpts = hooks.HookExecOpts{
					Hook:   r.clients.SDKConfig.Hooks.Start,
					Stdin:  bytes.NewBuffer(body),
					Stdout: r.clients.IO.WriteSecondary(r.clients.IO.WriteOut()),
					Stderr: r.clients.IO.WriteSecondary(r.clients.IO.WriteErr()),
				}

				out, err := r.clients.HookExecutor.Execute(ctx, startHookOpts)

				if err != nil {
					// Log the error but do not return because the user may be able to recover inside their app code
					r.log.Data["cloud_run_connection_command_error"] = fmt.Sprintf("%s\n%s", err, out)
					r.log.Warn("on_cloud_run_connection_command_error")
					break
				}

				r.log.Info("on_cloud_run_connection_command_output")

				linkResponse = &LinkResponse{
					EnvelopeID: msg.EnvelopeID,
					Payload:    json.RawMessage(out),
				}
			}

			// Write response back to websocket
			if linkResponse != nil {
				if err := sendWebSocketMessage(r.Connection, linkResponse); err != nil {
					errChan <- err
					return
				}
			}
		}
	}
}

// stopDelegateProcess terminates the currently running delegated process if one exists
func (r *LocalServer) stopDelegateProcess(ctx context.Context) {
	r.delegateCmdMutex.Lock()
	defer r.delegateCmdMutex.Unlock()

	if r.delegateCmd != nil {
		process := r.delegateCmd.GetProcess()
		if process != nil {
			r.clients.IO.PrintDebug(ctx, "Stopping previous delegated process (PID: %d)", process.Pid)
			// Kill the process gracefully
			err := process.Signal(os.Interrupt)
			if err != nil {
				// If interrupt fails, force kill
				r.clients.IO.PrintDebug(ctx, "Failed to interrupt process, sending SIGKILL: %v", err)
				_ = process.Kill()
			}
			// Wait for process to exit (with timeout)
			done := make(chan error)
			go func() {
				done <- r.delegateCmd.Wait()
			}()
			select {
			case <-done:
				r.clients.IO.PrintDebug(ctx, "Previous process stopped successfully")
			case <-time.After(5 * time.Second):
				r.clients.IO.PrintDebug(ctx, "Process did not exit in time, force killing")
				_ = process.Kill()
			}
		}
		r.delegateCmd = nil
	}
}

// StartDelegate passes along required opts to SDK, delegating
// connection for running app locally to script hook start
func (r *LocalServer) StartDelegate(ctx context.Context) error {
	// Set up hook execution options
	var sdkManagedConnectionStartHookOpts = hooks.HookExecOpts{
		Env: map[string]string{
			"SLACK_CLI_XAPP": r.token,
			"SLACK_CLI_XOXB": r.localHostedContext.BotAccessToken,
		},
		Exec: hooks.ShellExec{},
		Hook: r.clients.SDKConfig.Hooks.Start,
	}

	// Check whether hook script is available
	if !r.clients.SDKConfig.Hooks.Start.IsAvailable() {
		return slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The command for '%s' was not found", r.clients.SDKConfig.Hooks.Start.Name)
	}
	cmdStr, err := r.clients.SDKConfig.Hooks.Start.Get()
	if err != nil {
		return slackerror.New(slackerror.ErrSDKHookNotFound).WithRootCause(err)
	}

	// We're taking the script and separating it into individual fields to be compatible with Exec.Command,
	// then appending any additional arguments as flag --key=value pairs.
	cmdArgs := strings.Fields(cmdStr)
	var cmdArgVars = cmdArgs[1:] // omit the first item because that is the command name

	// Whatever cmd.Env is set to will be the ONLY environment variables that the `cmd` will have access to when it runs.
	// To avoid removing any environment variables that are set in the current environment, we first set the cmd.Env to the current environment.
	// before adding any new environment variables.
	var cmdEnvVars = os.Environ()
	cmdEnvVars = append(cmdEnvVars, goutils.MapToStringSlice(sdkManagedConnectionStartHookOpts.Env, "")...)
	cmd := sdkManagedConnectionStartHookOpts.Exec.Command(cmdEnvVars, os.Stdout, os.Stderr, nil, cmdArgs[0], cmdArgVars...)

	// Store command reference for lifecycle management
	r.delegateCmdMutex.Lock()
	r.delegateCmd = cmd
	r.delegateCmdMutex.Unlock()

	// Start the process (non-blocking)
	if err := cmd.Start(); err != nil {
		return slackerror.Wrap(err, slackerror.ErrSDKHookInvocationFailed).
			WithMessage("Failed to start 'start' hook")
	}

	// The following command will block, as the expectation is that SDK-delegated local-run invokes a long-running (blocking) child process
	err = cmd.Wait()
	if err != nil {
		if status, ok := err.(*exec.ExitError); ok {
			switch status.ExitCode() {
			case -1:
				return slackerror.New(slackerror.ErrProcessInterrupted)
			default:
				if status, ok := err.(*exec.ExitError); ok {
					code := iostreams.ExitCode(status.ExitCode())
					r.clients.IO.SetExitCode(code)
				}
				return slackerror.New(slackerror.ErrSDKHookInvocationFailed).
					WithMessage("The 'start' hook exited with an error").
					WithDetails(slackerror.ErrorDetails{
						{Code: slackerror.ErrLocalAppRun, Message: err.Error()},
					}).
					WithRemediation("")
			}
		}
		return err
	}
	return nil
}

// WatchManifest watches for manifest file changes and triggers app reinstallation
func (r *LocalServer) WatchManifest(ctx context.Context, auth types.SlackAuth, app types.App) error {
	// Check for watch SDKCLI configuration
	if !r.cliConfig.Config.Watch.IsAvailable() {
		r.clients.IO.PrintDebug(ctx, "To watch file changes, provide watch configuration in %s", config.GetProjectHooksJSONFilePath())
		return nil
	}

	// Get manifest watch configuration
	paths, filterRegex, enabled := r.cliConfig.Config.Watch.GetManifestWatchConfig()
	if !enabled {
		r.clients.IO.PrintDebug(ctx, "Manifest watching is not enabled")
		return nil
	}

	// Init watcher
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	// Use SDK-provided filter regex
	if filterRegex != "" {
		r.clients.IO.PrintDebug(ctx, "Watching manifest changes to file paths matching: %s", filterRegex)
		w.AddFilterHook(watcher.RegexFilterHook(regexp.MustCompile(filterRegex), false))
	}

	// Add provided paths to watcher
	for _, path := range paths {
		if err := w.AddRecursive(path); err != nil {
			r.log.Data["cloud_run_watch_error"] = fmt.Sprintf("manifest_watcher.paths: %s", err)
			r.log.Warn("on_cloud_run_watch_error")
		}
	}

	// Begin watching for manifest file changes
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.clients.IO.PrintDebug(ctx, "Manifest file watcher context canceled, returning.")
				return
			case event := <-w.Event:
				r.log.Data["cloud_run_watch_manifest_change"] = event.Path
				r.log.Info("on_cloud_run_watch_manifest_change")

				// Reinstall the app when manifest changes
				if _, _, _, err := apps.InstallLocalApp(ctx, r.clients, "", r.log, auth, app); err != nil {
					r.log.Data["cloud_run_watch_error"] = err.Error()
					r.log.Warn("on_cloud_run_watch_error")
				} else {
					r.log.Info("on_cloud_run_watch_manifest_change_reinstalled")
				}
			case err := <-w.Error:
				r.log.Data["cloud_run_watch_error"] = err.Error()
				r.log.Warn("on_cloud_run_watch_error")
			case <-w.Closed:
				return
			}
		}
	}()

	return w.Start(time.Millisecond * 100)
}

// WatchApp starts the delegated server and watches for app/code file changes to trigger restarts (SDK-managed connections only)
func (r *LocalServer) WatchApp(ctx context.Context) error {
	// Only run for SDK-managed connections
	if !r.cliConfig.Config.SDKManagedConnection {
		r.clients.IO.PrintDebug(ctx, "App watching is only enabled for SDK-managed connections")
		return nil
	}

	// Check for watch SDKCLI configuration
	watchAvailable := r.cliConfig.Config.Watch.IsAvailable()
	paths, filterRegex, watchEnabled := []string{}, "", false
	if watchAvailable {
		paths, filterRegex, watchEnabled = r.cliConfig.Config.Watch.GetAppWatchConfig()
	}

	// Start the initial delegated server process
	r.clients.IO.PrintDebug(ctx, "Starting initial delegated server process")
	serverErrChan := make(chan error, 1)
	go func() {
		err := r.StartDelegate(ctx)
		serverErrChan <- err
	}()

	// If watch is not configured or not enabled, just wait for the server to exit
	if !watchAvailable || !watchEnabled {
		r.clients.IO.PrintDebug(ctx, "App file watching is not enabled, running server without restart capability")
		return <-serverErrChan
	}

	// Init watcher for restarts
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	// Use SDK-provided filter regex
	if filterRegex != "" {
		r.clients.IO.PrintDebug(ctx, "Watching app changes to file paths matching: %s", filterRegex)
		w.AddFilterHook(watcher.RegexFilterHook(regexp.MustCompile(filterRegex), false))
	}

	// Add provided paths to watcher
	for _, path := range paths {
		if err := w.AddRecursive(path); err != nil {
			r.log.Data["cloud_run_watch_error"] = fmt.Sprintf("app_watcher.paths: %s", err)
			r.log.Warn("on_cloud_run_watch_error")
		}
	}

	// Begin watching for app file changes
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.clients.IO.PrintDebug(ctx, "App file watcher context canceled, returning.")
				return
			case event := <-w.Event:
				r.log.Data["cloud_run_watch_app_change"] = event.Path
				r.log.Info("on_cloud_run_watch_app_change")

				// Stop the previous process before starting a new one
				r.stopDelegateProcess(ctx)

				// Start new delegated server process
				go func() {
					err := r.StartDelegate(ctx)
					if err != nil {
						r.clients.IO.PrintDebug(ctx, "Delegated start hook failed on restart: %v", err)
						return
					}
				}()
			case err := <-w.Error:
				r.log.Data["cloud_run_watch_error"] = err.Error()
				r.log.Warn("on_cloud_run_watch_error")
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watcher and wait for either the watcher or server to error
	if err := w.Start(time.Millisecond * 100); err != nil {
		return err
	}

	// Wait for the initial server to exit (if it does)
	return <-serverErrChan
}

func (r *LocalServer) WatchActivityLogs(ctx context.Context, minLevel string) error {
	// Default minimum log level
	if strings.TrimSpace(minLevel) == "" {
		minLevel = ActivityMinLevelDefault
	}

	var activityArgs = types.ActivityArgs{
		TeamID:            r.localHostedContext.TeamID,
		AppID:             r.localHostedContext.AppID,
		TailArg:           true,
		PollingIntervalMS: ActivityPollingIntervalDefault * 1000,
		MinDateCreated:    time.Now().UnixMicro(),
		MinLevel:          minLevel,
		Limit:             ActivityLimitDefault,

		// Timeout after 24 hours - TODO(@mbrooks) can we remove the timeout entirely?
		IdleTimeoutM: 60 * 24,
	}
	// Next line runs in a ticker loop (based on TailArg above) that will return if the context is cancelled or an error occurs
	return Activity(ctx, r.clients, r.log, activityArgs)
}

// Message describes a web socket incoming message
type Message struct {
	Type                   string          `json:"type,omitempty"`
	DebugInfo              DebugInfo       `json:"debug_info,omitempty"`
	Reason                 string          `json:"reason,omitempty"`
	EnvelopeID             string          `json:"envelope_id,omitempty"`
	Payload                json.RawMessage `json:"payload,omitempty"`
	AcceptsResponsePayload bool            `json:"accepts_response_payload,omitempty"`
}

// DebugInfo may be included in the web socket's incoming Message
type DebugInfo struct {
	Host    string `json:"host,omitempty"`
	Started string `json:"started,omitempty"`
	Build   int    `json:"build,omitempty"`
}

// LinkResponse describes a web socket response
type LinkResponse struct {
	EnvelopeID string          `json:"envelope_id,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

// Constants for web socket message types
const (
	helloMessageType      string = "hello"
	disconnectMessageType string = "disconnect"

	// eventsAPIMessageType     string = "events_api"
	// slashCommandMessageType  string = "slash_command"
	// interactivityMessageType string = "interactive"

	// UNRECOVERABLE bool = false
	// RECOVERABLE   bool = true
)

// SocketEvent describes an incoming socket event for the SDK
type SocketEvent struct {
	Body    json.RawMessage    `json:"body,omitempty"`
	Context LocalHostedContext `json:"context,omitempty"`
}

// sendWebSocketMessage marshal's the linkResponse to JSON and sends it as a TextMessage type using the provided websocket connection (c).
func sendWebSocketMessage(c WebSocketConnection, linkResponse *LinkResponse) error {
	// Validate the response
	if linkResponse == nil {
		return slackerror.Wrap(fmt.Errorf("websocket response message cannot be empty"), slackerror.ErrSocketConnection)
	}

	// Prepare response for websocket
	b, err := json.Marshal(*linkResponse)
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrSocketConnection)
	}

	// Write response back to websocket
	err = c.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrSocketConnection)
	}

	return nil
}
