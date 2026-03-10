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

package mailencoding

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// DecodeRFC2047 は RFC 2047 形式のエンコードされたテキストをデコードします
// 例: "=?UTF-8?B?44GE44KT44G744Gm44GE44KT?=" -> "日本語"
func DecodeRFC2047(encoded string) string {
	if !strings.Contains(encoded, "=?") {
		return encoded
	}

	// RFC 2047 形式にマッチするパターン
	// =?charset?encoding?encoded-text?=
	pattern := regexp.MustCompile(`=\?([^?]+)\?([^?]+)\?([^?]+)\?=`)
	matches := pattern.FindAllStringSubmatch(encoded, -1)

	if matches == nil {
		return encoded
	}

	result := encoded
	for _, match := range matches {
		fullMatch := match[0]
		charset := strings.ToUpper(match[1])
		encoding := strings.ToUpper(match[2])
		encodedText := match[3]

		decodedText := decodeRFC2047Part(charset, encoding, encodedText)
		result = strings.ReplaceAll(result, fullMatch, decodedText)
	}

	return result
}

func decodeRFC2047Part(charset, encoding, encodedText string) string {
	var decodedBytes []byte
	var err error

	// デコード処理（Base64またはQuoted-Printable）
	switch encoding {
	case "B":
		decodedBytes, err = base64.StdEncoding.DecodeString(encodedText)
		if err != nil {
			return encodedText
		}
	case "Q":
		reader := quotedprintable.NewReader(strings.NewReader(encodedText))
		decodedBytes, err = io.ReadAll(reader)
		if err != nil {
			return encodedText
		}
	default:
		return encodedText
	}

	// 文字コードの変換
	decoder := getDecoder(charset)
	if decoder != nil {
		utf8Bytes, err := decodeString(decodedBytes, decoder)
		if err == nil {
			return string(utf8Bytes)
		}
	}

	return string(decodedBytes)
}

func getDecoder(charset string) transform.Transformer {
	switch strings.ToUpper(charset) {
	case "UTF-8", "UTF8":
		return unicode.UTF8.NewDecoder()
	case "ISO-2022-JP":
		return japanese.ISO2022JP.NewDecoder()
	case "SHIFT_JIS", "SHIFT-JIS", "SJIS":
		return japanese.ShiftJIS.NewDecoder()
	case "EUC-JP":
		return japanese.EUCJP.NewDecoder()
	case "GB2312", "GBK":
		return simplifiedchinese.GBK.NewDecoder()
	case "BIG5":
		return traditionalchinese.Big5.NewDecoder()
	default:
		return nil
	}
}

func decodeString(b []byte, d transform.Transformer) ([]byte, error) {
	return transform.Bytes(d, b)
}
