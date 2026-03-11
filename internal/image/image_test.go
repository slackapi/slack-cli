package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

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

func TestResizeImage(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	reader := bytes.NewReader(pngData)

	resized, err := ResizeImage(reader, 50, 50)
	require.NoError(t, err)
	assert.NotNil(t, resized)
	assert.Equal(t, 50, resized.Bounds().Dx())
	assert.Equal(t, 50, resized.Bounds().Dy())
}

func TestResizeImageToBytes(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	reader := bytes.NewReader(pngData)

	result, err := ResizeImageToBytes(reader, 50, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestResizeImageFromFileToBytes(t *testing.T) {
	pngData := createTestPNG(t, 100, 100)
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := ResizeImageFromFileToBytes(fs, "/test.png", 50, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestResizeImageFromFileToBytes_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := ResizeImageFromFileToBytes(fs, "/nonexistent.png", 50, 50)
	assert.Error(t, err)
}

func TestCropResizeImageRatio(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	reader := bytes.NewReader(pngData)

	result, err := CropResizeImageRatio(reader, 100, 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result.Bounds().Dx())
}

func TestCropResizeImageRatioToBytes(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	reader := bytes.NewReader(pngData)

	result, err := CropResizeImageRatioToBytes(reader, 100, 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestCropResizeImageRatioFromFile(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := CropResizeImageRatioFromFile(fs, "/test.png", 100, 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCropResizeImageRatioFromFile_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := CropResizeImageRatioFromFile(fs, "/nonexistent.png", 100, 1, 1)
	assert.Error(t, err)
}

func TestCropResizeImageRatioFromFileToBytes(t *testing.T) {
	pngData := createTestPNG(t, 200, 100)
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/test.png", pngData, 0644)
	require.NoError(t, err)

	result, err := CropResizeImageRatioFromFileToBytes(fs, "/test.png", 100, 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestCropResizeImageRatioFromFileToBytes_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := CropResizeImageRatioFromFileToBytes(fs, "/nonexistent.png", 100, 1, 1)
	assert.Error(t, err)
}

func TestResizeImage_InvalidReader(t *testing.T) {
	reader := bytes.NewReader([]byte("not a valid image"))
	_, err := ResizeImage(reader, 50, 50)
	assert.Error(t, err)
}
