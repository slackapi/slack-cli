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

package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"runtime"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/contextutil"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// UploadPackageToS3 uploads an archive file to S3, given the provided upload params.
// TODO: Should be moved out of Client, since it just reuses the httpclient instance directly
func (c *Client) UploadPackageToS3(ctx context.Context, fs afero.Fs, appID string, uploadParams GenerateS3PresignedPostResult, archiveFilePath string) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.UploadPackageToS3")
	span.SetTag("app", appID)
	defer span.Finish()

	extraParams := map[string]string{
		"X-Amz-Credential":     uploadParams.Fields.AmzCredentials,
		"X-Amz-Algorithm":      uploadParams.Fields.AmzAlgorithm,
		"key":                  uploadParams.Fields.AmzFileKey,
		"X-Amz-Date":           uploadParams.Fields.AmzFileCreateDate,
		"Policy":               uploadParams.Fields.AmzPolicy,
		"X-Amz-Signature":      uploadParams.Fields.AmzSignature,
		"X-Amz-Security-Token": uploadParams.Fields.AmzToken,
	}

	fileName := uploadParams.FileName

	if archiveFilePath == "" || appID == "" {
		return uploadParams.FileName, slackerror.New("missing required args")
	}
	archive, err := fs.Open(archiveFilePath)
	if err != nil {
		return fileName, err
	}
	defer archive.Close()

	archiveBytes, err := io.ReadAll(archive)
	if err != nil {
		return fileName, err
	}

	var uploadbody = new(bytes.Buffer)
	var writer = multipart.NewWriter(uploadbody)

	for key, val := range extraParams {
		err = writer.WriteField(key, val)
		if err != nil {
			return fileName, err
		}
	}

	md5hash := md5.New()
	if _, err := io.Copy(md5hash, archive); err != nil {
		return fileName, err
	}

	md5s := base64.StdEncoding.EncodeToString(md5hash.Sum(nil))

	var part io.Writer
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileName))
	h.Set("Content-Type", "application/zip")
	part, err = writer.CreatePart(h)
	if err != nil {
		return fileName, err
	}
	_, err = part.Write(archiveBytes)
	if err != nil {
		return fileName, err
	}

	// close the writer after the body is formed
	writer.Close()

	var request *http.Request
	request, err = http.NewRequestWithContext(ctx, "POST", uploadParams.Url, uploadbody)
	if err != nil {
		return fileName, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Content-MD5", md5s)
	cliVersion := contextutil.VersionFromContext(ctx)
	var userAgent = fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)
	request.Header.Add("User-Agent", userAgent)

	var s3span = opentracing.StartSpan("apiclient.UploadPackageToS3.FileUpload", opentracing.ChildOf(span.Context()))
	s3span.SetTag("app", appID)
	defer s3span.Finish()
	uploadresp, err := c.httpClient.Do(request)
	if err != nil {
		return fileName, err
	}

	defer uploadresp.Body.Close()
	if uploadresp.StatusCode != http.StatusNoContent {
		return fileName, slackerror.New("Failed uploads to s3")
	}

	return fileName, nil
}
