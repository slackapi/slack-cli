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
	"context"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var imgFile = ".assets/icon.png"

func TestClient_IconErrorIfMissingArgs(t *testing.T) {
	fs := afero.NewMemMapFs()
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appIconMethod,
	})
	defer teardown()
	_, err := c.Icon(context.Background(), fs, "token", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required args")
}

func TestClient_IconErrorNoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appIconMethod,
	})
	defer teardown()
	_, err := c.Icon(context.Background(), fs, "token", "12345", imgFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "file does not exist")
}

func TestClient_IconErrorWrongFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "test.txt", []byte("this is a text file"), 0666)
	require.NoError(t, err)
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appIconMethod,
	})
	defer teardown()
	_, err = c.Icon(context.Background(), fs, "token", "12345", "test.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown format")
}

func TestClient_IconSuccess(t *testing.T) {
	fs := afero.NewMemMapFs()

	myimage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}})

	// This loop just fills the image with random data
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			c := color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
			myimage.Set(x, y, c)
		}
	}
	myfile, _ := fs.Create(imgFile)
	err := png.Encode(myfile, myimage)
	require.NoError(t, err)
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appIconMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	_, err = c.Icon(context.Background(), fs, "token", "12345", imgFile)
	require.NoError(t, err)
}
