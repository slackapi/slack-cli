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

package image

// borrowed from Missions

import (
	"bytes"
	"image"
	"io"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/spf13/afero"
)

func CropResizeImageRatioFromFile(fs afero.Fs, filepath string, width uint, widthRatio, heightRatio int) (image.Image, error) {
	reader, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return CropResizeImageRatio(reader, width, widthRatio, heightRatio)
}

func CropResizeImageRatioFromFileToBytes(fs afero.Fs, filepath string, width uint, widthRatio, heightRatio int) ([]byte, error) {
	reader, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return CropResizeImageRatioToBytes(reader, width, widthRatio, heightRatio)
}

func CropResizeImageRatio(reader io.Reader, width uint, widthRatio, heightRatio int) (image.Image, error) {
	originalImage, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	croppedImg, err := cutter.Crop(originalImage, cutter.Config{
		Width:   widthRatio,
		Height:  heightRatio,
		Mode:    cutter.Centered,
		Options: cutter.Ratio,
	})
	if err != nil {
		return nil, err
	}

	return resize.Resize(width, 0, croppedImg, resize.Lanczos3), nil
}

func CropResizeImageRatioToBytes(reader io.Reader, width uint, widthRatio, heightRatio int) ([]byte, error) {
	resizedImg, err := CropResizeImageRatio(reader, width, widthRatio, heightRatio)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, resizedImg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ResizeImage(reader io.Reader, width, height uint) (image.Image, error) {
	originalImage, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return resize.Resize(width, height, originalImage, resize.Lanczos3), nil
}

func ResizeImageToBytes(reader io.Reader, width, height uint) ([]byte, error) {
	resizedImg, err := ResizeImage(reader, width, height)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, resizedImg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ResizeImageFromFileToBytes(fs afero.Fs, filepath string, width, height uint) ([]byte, error) {
	reader, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ResizeImageToBytes(reader, width, height)
}
