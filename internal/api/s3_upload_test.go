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
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestClient_UploadPackageToS3(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "foo.txt", []byte("this is the package"), 0666)
	require.NoError(t, err)

	creds := "creds"
	algo := "algo"
	key := "key"
	createDate := "2022-01-01"
	policy := "policy"
	signature := "sig"
	token := "token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, "/s3upload")

		contentType := r.Header["Content-Type"][0]
		_, contentTypeParams, err := mime.ParseMediaType(contentType)
		require.NoError(t, err)
		boundary := contentTypeParams["boundary"]
		form, err := multipart.NewReader(r.Body, boundary).ReadForm(1024)
		require.NoError(t, err)
		require.Equal(t, creds, form.Value["X-Amz-Credential"][0])
		require.Equal(t, algo, form.Value["X-Amz-Algorithm"][0])
		require.Equal(t, key, form.Value["key"][0])
		require.Equal(t, createDate, form.Value["X-Amz-Date"][0])
		require.Equal(t, policy, form.Value["Policy"][0])
		require.Equal(t, signature, form.Value["X-Amz-Signature"][0])
		require.Equal(t, token, form.Value["X-Amz-Security-Token"][0])

		file := form.File["file"][0]
		require.NotNil(t, file)
		require.Equal(t, "foo.txt", file.Filename)
		content, err := file.Open()
		require.NoError(t, err)
		contentBytes, err := io.ReadAll(content)
		require.NoError(t, err)
		require.Equal(t, "this is the package", string(contentBytes))

		md5Header := r.Header["Content-Md5"][0]
		require.Equal(t, "1B2M2Y8AsgTpgAmY7PhCfg==", md5Header)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := NewClient(&http.Client{}, server.URL, nil)

	s3Params := GenerateS3PresignedPostResult{
		Url:      server.URL + "/s3upload",
		FileName: "foo.txt",
		Fields: PresignedPostFields{
			AmzCredentials:    creds,
			AmzAlgorithm:      algo,
			AmzFileKey:        key,
			AmzFileCreateDate: createDate,
			AmzPolicy:         policy,
			AmzSignature:      signature,
			AmzToken:          token,
		},
	}
	resp, err := client.UploadPackageToS3(ctx, fs, "appID", s3Params, "foo.txt")
	require.NoError(t, err)
	require.Equal(t, "foo.txt", resp)
}
