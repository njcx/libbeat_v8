// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sys

import (
	"strings"

	"github.com/njcx/libbeat_v8/common"
)

// UTF16BytesToString converts the given UTF-16 bytes to a string.
func UTF16BytesToString(b []byte) (string, error) {
	// Use space from the ByteBuffer pool as working memory for the conversion.
	bb := NewPooledByteBuffer()
	defer bb.Free()

	if err := common.UTF16ToUTF8Bytes(b, bb); err != nil {
		return "", err
	}

	// This copies the UTF-8 bytes to create a string.
	return string(bb.Bytes()), nil
}

// RemoveWindowsLineEndings replaces carriage return line feed (CRLF) with
// line feed (LF) and trims any newline character that may exist at the end
// of the string.
func RemoveWindowsLineEndings(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	return strings.TrimRight(s, "\n")
}

// BinaryToString converts a binary field which is encoded in hexadecimal
// to its string representation. This is equivalent to hex.EncodeToString
// but its output is in uppercase to be equivalent to the windows
// XML formatting of this fields.
func BinaryToString(bin []byte) string {
	if len(bin) == 0 {
		return ""
	}

	const hexTable = "0123456789ABCDEF"

	size := len(bin) * 2
	buffer := make([]byte, size)

	j := 0
	for _, v := range bin {
		buffer[j] = hexTable[v>>4]
		buffer[j+1] = hexTable[v&0x0f]
		j += 2
	}

	return string(buffer)
}
