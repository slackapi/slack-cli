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

package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := range width {
		for y := range height {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)
	return buf.Bytes()
}

func Test_ResizeImage(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	reader := bytes.NewReader(pngData)

	resized, err := ResizeImage(reader, 50, 50)
	require.NoError(t, err)
	assert.NotNil(t, resized)
	assert.Equal(t, 50, resized.Bounds().Dx())
	assert.Equal(t, 50, resized.Bounds().Dy())
}

func Test_ResizeImageToBytes(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	reader := bytes.NewReader(pngData)

	result, err := ResizeImageToBytes(reader, 50, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func Test_ResizeImageFromFileToBytes(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	fs := slackdeps.NewFsMock()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := ResizeImageFromFileToBytes(fs, "/test.png", 50, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func Test_ResizeImageFromFileToBytes_FileNotFound(t *testing.T) {
	fs := slackdeps.NewFsMock()
	_, err := ResizeImageFromFileToBytes(fs, "/nonexistent.png", 50, 50)
	assert.Error(t, err)
}

func Test_CropResizeImageRatio(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	reader := bytes.NewReader(pngData)

	result, err := CropResizeImageRatio(reader, 100, 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result.Bounds().Dx())
}

func Test_CropResizeImageRatioToBytes(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	reader := bytes.NewReader(pngData)

	result, err := CropResizeImageRatioToBytes(reader, 100, 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func Test_CropResizeImageRatioFromFile(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	fs := slackdeps.NewFsMock()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := CropResizeImageRatioFromFile(fs, "/test.png", 100, 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_CropResizeImageRatioFromFile_FileNotFound(t *testing.T) {
	fs := slackdeps.NewFsMock()
	_, err := CropResizeImageRatioFromFile(fs, "/nonexistent.png", 100, 1, 1)
	assert.Error(t, err)
}

func Test_CropResizeImageRatioFromFileToBytes(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	fs := slackdeps.NewFsMock()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := CropResizeImageRatioFromFileToBytes(fs, "/test.png", 100, 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func Test_CropResizeImageRatioFromFileToBytes_FileNotFound(t *testing.T) {
	fs := slackdeps.NewFsMock()
	_, err := CropResizeImageRatioFromFileToBytes(fs, "/nonexistent.png", 100, 1, 1)
	assert.Error(t, err)
}

func Test_ResizeImage_InvalidReader(t *testing.T) {
	reader := bytes.NewReader([]byte("not a valid image"))
	_, err := ResizeImage(reader, 50, 50)
	assert.Error(t, err)
}
