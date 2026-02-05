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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"runtime"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/image"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

const (
	appIconMethod = "apps.hosted.icon"
)

// IconResult details to be saved
type IconResult struct {
}

type iconResponse struct {
	extendedBaseResponse
	IconResult
}

// Icon updates a Slack App's icon
func (c *Client) Icon(ctx context.Context, fs afero.Fs, token, appID, iconFilePath string) (IconResult, error) {
	var (
		iconBytes []byte
		err       error
		span      opentracing.Span
	)
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.Icon")
	defer span.Finish()

	if iconFilePath == "" || appID == "" {
		return IconResult{}, slackerror.New("missing required args")
	}

	icon, err := fs.Open(iconFilePath)
	if err != nil {
		return IconResult{}, err
	}
	defer icon.Close()

	iconBytes, err = image.CropResizeImageRatioFromFileToBytes(fs, iconFilePath, 512, 1, 1)
	if err != nil {
		return IconResult{}, err
	}

	iconStat, err := icon.Stat()
	if err != nil {
		return IconResult{}, err
	}

	var body = new(bytes.Buffer)
	var writer = multipart.NewWriter(body)

	var part io.Writer
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", iconStat.Name()))
	h.Set("Content-Type", http.DetectContentType(iconBytes))
	part, err = writer.CreatePart(h)
	if err != nil {
		return IconResult{}, err
	}

	_, err = part.Write(iconBytes)
	if err != nil {
		return IconResult{}, err
	}

	err = writer.WriteField("app_id", appID)
	if err != nil {
		return IconResult{}, err
	}
	// close the writer after the body is formed
	writer.Close()

	var sURL *url.URL
	sURL, err = url.Parse(c.host + "/api/" + appIconMethod)
	if err != nil {
		return IconResult{}, err
	}

	span.SetTag("request_url", sURL)

	var request *http.Request
	request, err = http.NewRequestWithContext(ctx, "POST", sURL.String(), body)
	if err != nil {
		return IconResult{}, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", "Bearer "+token)
	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return IconResult{}, err
	}
	var userAgent = fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)
	request.Header.Add("User-Agent", userAgent)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return IconResult{}, err
	}
	defer resp.Body.Close()

	span.SetTag("status_code", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return IconResult{}, err
	}

	var result iconResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return IconResult{}, err
	}

	span.SetTag("ok", result.Ok)

	if !result.Ok {
		span.SetTag("error", result.Error)
		return IconResult{}, fmt.Errorf("%s error: %s", sURL.String(), result.Error)
	}

	// return result
	return IconResult{}, nil
}
