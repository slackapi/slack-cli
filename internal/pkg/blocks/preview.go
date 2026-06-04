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
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

var connectionTimeout = 30 * time.Second
var responseTimeout = 30 * time.Second

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

func openInBrowser(ctx context.Context, io iostreams.IOStreamer, browser slackdeps.Browser, url string) {
	io.PrintDebug(ctx, "Opening Block Kit Builder: %s", url)
	io.PrintInfo(ctx, false, "Opening Block Kit Builder in your browser...")
	browser.OpenURL(url)
}

func handshake(ctx context.Context, io iostreams.IOStreamer, ws wsConn) error {
	msg, err := ws.readMessage(responseTimeout)
	if err != nil {
		return err
	}
	if msg.Type != "CONNECTED" {
		return slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Unexpected message type: %s", msg.Type)
	}
	var cp connectedPayload
	if err := json.Unmarshal(msg.Payload, &cp); err == nil {
		io.PrintDebug(ctx, "Block Kit Builder version: %s", cp.Version)
	}
	return nil
}

func requestScreenshot(ctx context.Context, io iostreams.IOStreamer, ws wsConn) (screenshotPayload, error) {
	if err := ws.writeMessage(wsMessage{Type: "REQUEST_SCREENSHOT"}); err != nil {
		return screenshotPayload{}, err
	}
	io.PrintDebug(ctx, "Sent REQUEST_SCREENSHOT")

	response, err := ws.readMessage(responseTimeout)
	if err != nil {
		return screenshotPayload{}, err
	}

	switch response.Type {
	case "SCREENSHOT":
		var sp screenshotPayload
		if err := json.Unmarshal(response.Payload, &sp); err != nil {
			return screenshotPayload{}, err
		}
		return sp, nil
	case "ERROR":
		var ep errorPayload
		if err := json.Unmarshal(response.Payload, &ep); err != nil {
			return screenshotPayload{}, slackerror.New(slackerror.ErrBlocksPreview).
				WithMessage("Block Kit Builder returned an error")
		}
		return screenshotPayload{}, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Block Kit Builder error: %s", ep.Message)
	default:
		return screenshotPayload{}, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Unexpected response type: %s", response.Type)
	}
}

// decodeImage parses a RFC 2397 data URL (e.g. "data:image/png;base64,...")
// and returns the decoded binary data. Only base64-encoded data URLs are supported.
func decodeImage(dataURL string) ([]byte, error) {
	after, found := strings.CutPrefix(dataURL, "data:")
	if !found {
		return nil, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Invalid image data: missing data URL scheme")
	}

	metadata, data, found := strings.Cut(after, ",")
	if !found {
		return nil, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Invalid image data: missing comma separator in data URL")
	}

	if !strings.HasSuffix(metadata, ";base64") {
		return nil, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Invalid image data: data URL is not base64-encoded")
	}

	return base64.StdEncoding.DecodeString(data)
}

func saveImage(fs afero.Fs, outputPath string, data []byte) error {
	if err := fs.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	return afero.WriteFile(fs, outputPath, data, 0644)
}

func Preview(ctx context.Context, clients *shared.ClientFactory, teamID string, blocksJSON string, outputPath string) (string, error) {
	wsServer, err := newWebSocketServer()
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	defer wsServer.Shutdown()

	builderURL, err := buildBlockKitBuilderURL(clients.API().Host(), teamID, wsServer.Port, blocksJSON)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	openInBrowser(ctx, clients.IO, clients.Browser(), builderURL)

	ws, err := wsServer.Accept(ctx, connectionTimeout)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}
	defer ws.Close()
	clients.IO.PrintInfo(ctx, false, "Block Kit Builder connected")

	if err := handshake(ctx, clients.IO, ws); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	screenshot, err := requestScreenshot(ctx, clients.IO, ws)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	imageBytes, err := decodeImage(screenshot.Image)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	if err := saveImage(clients.Fs, outputPath, imageBytes); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrBlocksPreview)
	}

	clients.IO.PrintDebug(ctx, "Screenshot saved: %s (%dx%d)", outputPath, screenshot.Width, screenshot.Height)
	return outputPath, nil
}

func buildBlockKitBuilderURL(apiHost string, teamID string, port int, blocksJSON string) (string, error) {
	parsed, err := url.Parse(apiHost)
	if err != nil {
		return "", fmt.Errorf("invalid API host %q: %w", apiHost, err)
	}
	parsed.Host = "app." + parsed.Host
	parsed.Path = fmt.Sprintf("/block-kit-builder/%s/builder", teamID)
	q := parsed.Query()
	q.Set("ws_port", fmt.Sprintf("%d", port))
	parsed.RawQuery = q.Encode()
	parsed.Fragment = blocksJSON
	return parsed.String(), nil
}
