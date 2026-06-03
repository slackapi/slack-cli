// Copyright 2022-2026 Salesforce, Inc.
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

package blocks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

var connectionTimeout = 30 * time.Second
var responseTimeout = 30 * time.Second

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var netListen = net.Listen

const blockKitBuilderURLTemplate = "https://app.dev1388.slack.com/block-kit-builder/%s/builder"

type wsMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type screenshotPayload struct {
	Image  string `json:"image"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type errorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type connectedPayload struct {
	Version string `json:"version"`
}

func readWSMessage(conn *websocket.Conn, timeout time.Duration) (wsMessage, error) {
	type result struct {
		msg wsMessage
		err error
	}
	ch := make(chan result, 1)
	go func() {
		_, data, err := conn.ReadMessage()
		if err != nil {
			ch <- result{err: err}
			return
		}
		var msg wsMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			ch <- result{err: fmt.Errorf("invalid message from Block Kit Builder: %w", err)}
			return
		}
		ch <- result{msg: msg}
	}()

	select {
	case r := <-ch:
		return r.msg, r.err
	case <-time.After(timeout):
		return wsMessage{}, fmt.Errorf("timed out waiting for message")
	}
}

func Preview(ctx context.Context, clients *shared.ClientFactory, teamID string, blocksJSON string, outputPath string) (string, error) {
	listener, err := netListen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	connChan := make(chan *websocket.Conn, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, upgradeErr := upgrader.Upgrade(w, r, nil)
		if upgradeErr != nil {
			errChan <- upgradeErr
			return
		}
		connChan <- conn
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Shutdown(context.Background()) }()

	builderURL := buildBlockKitBuilderURL(teamID, port, blocksJSON)
	clients.IO.PrintDebug(ctx, "Opening Block Kit Builder: %s", builderURL)
	clients.IO.PrintInfo(ctx, false, "Opening Block Kit Builder in your browser...")
	clients.Browser().OpenURL(builderURL)

	var conn *websocket.Conn
	select {
	case conn = <-connChan:
		clients.IO.PrintInfo(ctx, false, "Block Kit Builder connected")
	case err := <-errChan:
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	case <-time.After(connectionTimeout):
		return "", slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Timed out waiting for Block Kit Builder to connect")
	case <-ctx.Done():
		return "", slackerror.Wrap(ctx.Err(), slackerror.ErrBlocksPreview)
	}
	defer conn.Close()

	connectedMsg, err := readWSMessage(conn, responseTimeout)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	if connectedMsg.Type != "CONNECTED" {
		return "", slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Unexpected message type: %s", connectedMsg.Type)
	}
	var cp connectedPayload
	if err := json.Unmarshal(connectedMsg.Payload, &cp); err == nil {
		clients.IO.PrintDebug(ctx, "Block Kit Builder version: %s", cp.Version)
	}

	reqMsg := wsMessage{Type: "REQUEST_SCREENSHOT"}
	reqBytes, _ := json.Marshal(reqMsg)
	if err := conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	clients.IO.PrintDebug(ctx, "Sent REQUEST_SCREENSHOT")

	response, err := readWSMessage(conn, responseTimeout)
	if err != nil {
		return "", slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Timed out waiting for screenshot response")
	}

	if response.Type == "ERROR" {
		var ep errorPayload
		if err := json.Unmarshal(response.Payload, &ep); err != nil {
			return "", slackerror.New(slackerror.ErrBlocksPreview).
				WithMessage("Block Kit Builder returned an error")
		}
		return "", slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Block Kit Builder error: %s", ep.Message)
	}

	if response.Type != "SCREENSHOT" {
		return "", slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Unexpected response type: %s", response.Type)
	}

	var sp screenshotPayload
	if err := json.Unmarshal(response.Payload, &sp); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	imageBytes, err := base64.StdEncoding.DecodeString(sp.Image)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	if err := clients.Fs.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	if err := afero.WriteFile(clients.Fs, outputPath, imageBytes, 0644); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	clients.IO.PrintDebug(ctx, "Screenshot saved: %s (%dx%d)", outputPath, sp.Width, sp.Height)
	return outputPath, nil
}

func buildBlockKitBuilderURL(teamID string, port int, blocksJSON string) string {
	base := fmt.Sprintf(blockKitBuilderURLTemplate, teamID)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("ws_port", fmt.Sprintf("%d", port))
	u.RawQuery = q.Encode()
	u.Fragment = blocksJSON
	return u.String()
}
