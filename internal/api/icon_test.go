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
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var imgFile = ".assets/icon.png"

func createTestPNG(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	myimage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}})
	for x := range 100 {
		for y := range 100 {
			c := color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
			myimage.Set(x, y, c)
		}
	}
	myfile, err := fs.Create(path)
	require.NoError(t, err)
	err = png.Encode(myfile, myimage)
	require.NoError(t, err)
}

func TestClient_Icon(t *testing.T) {
	tests := map[string]struct {
		setupFs     func(t *testing.T, fs afero.Fs)
		appID       string
		filePath    string
		response    string
		expectErr   bool
		errContains string
	}{
		"returns error when args are missing": {
			appID:       "",
			filePath:    "",
			expectErr:   true,
			errContains: "missing required args",
		},
		"returns error when file does not exist": {
			appID:       "12345",
			filePath:    imgFile,
			expectErr:   true,
			errContains: "file does not exist",
		},
		"returns error for non-image file": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				err := afero.WriteFile(fs, "test.txt", []byte("this is a text file"), 0666)
				require.NoError(t, err)
			},
			appID:       "12345",
			filePath:    "test.txt",
			expectErr:   true,
			errContains: "unknown format",
		},
		"succeeds with valid PNG": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				createTestPNG(t, fs, imgFile)
			},
			appID:    "12345",
			filePath: imgFile,
			response: `{"ok":true}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := afero.NewMemMapFs()
			if tc.setupFs != nil {
				tc.setupFs(t, fs)
			}
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appIconMethod,
				Response:       tc.response,
			})
			defer teardown()
			_, err := c.Icon(ctx, fs, "token", tc.appID, tc.filePath)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_IconSet(t *testing.T) {
	tests := map[string]struct {
		setupFs     func(t *testing.T, fs afero.Fs)
		appID       string
		filePath    string
		response    string
		expectErr   bool
		errContains string
	}{
		"returns error when args are missing": {
			appID:       "",
			filePath:    "",
			expectErr:   true,
			errContains: "missing required args",
		},
		"returns error when file does not exist": {
			appID:       "12345",
			filePath:    imgFile,
			expectErr:   true,
			errContains: "file does not exist",
		},
		"returns error for empty file": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				err := afero.WriteFile(fs, imgFile, []byte{}, 0666)
				require.NoError(t, err)
			},
			appID:     "12345",
			filePath:  imgFile,
			expectErr: true,
		},
		"returns error for unsupported format": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				svgContent := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><circle cx="50" cy="50" r="50"/></svg>`)
				err := afero.WriteFile(fs, "icon.svg", svgContent, 0666)
				require.NoError(t, err)
			},
			appID:       "12345",
			filePath:    "icon.svg",
			expectErr:   true,
			errContains: "unknown format",
		},
		"returns error from API response": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				createTestPNG(t, fs, imgFile)
			},
			appID:       "12345",
			filePath:    imgFile,
			response:    `{"ok":false,"error":"invalid_app"}`,
			expectErr:   true,
			errContains: "invalid_app",
		},
		"succeeds with valid PNG": {
			setupFs: func(t *testing.T, fs afero.Fs) {
				createTestPNG(t, fs, imgFile)
			},
			appID:    "12345",
			filePath: imgFile,
			response: `{"ok":true}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := afero.NewMemMapFs()
			if tc.setupFs != nil {
				tc.setupFs(t, fs)
			}
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appIconSetMethod,
				Response:       tc.response,
			})
			defer teardown()
			_, err := c.IconSet(ctx, fs, "token", tc.appID, tc.filePath)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
