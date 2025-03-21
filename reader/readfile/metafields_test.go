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

package readfile

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/njcx/libbeat_v8/reader"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestMetaFields(t *testing.T) {
	messages := []reader.Message{
		{
			Content: []byte("my line"),
			Bytes:   7,
			Fields:  mapstr.M{},
		},
		{
			Content: []byte("my line again"),
			Bytes:   13,
			Fields:  mapstr.M{},
		},
		{
			Content: []byte(""),
			Bytes:   10,
			Fields:  mapstr.M{},
		},
	}

	path := "test/path"
	offset := int64(0)

	in := &FileMetaReader{msgReader(messages), path, createTestFileInfo(), "hash", offset}
	for {
		msg, err := in.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		expectedFields := mapstr.M{}
		if len(msg.Content) != 0 {
			expectedFields = mapstr.M{
				"log": mapstr.M{
					"file": mapstr.M{
						"path":        path,
						"fingerprint": "hash",
					},
					"offset": offset,
				},
			}
			checkFields(t, expectedFields, msg.Fields)
		} else {
			require.Equal(t, expectedFields, msg.Fields)
		}
		offset += int64(msg.Bytes)

		require.Equal(t, offset, in.offset)
	}
}

func msgReader(m []reader.Message) reader.Reader {
	return &messageReader{
		messages: m,
	}
}

type messageReader struct {
	messages []reader.Message
	i        int
}

func (r *messageReader) Next() (reader.Message, error) {
	if r.i == len(r.messages) {
		return reader.Message{}, io.EOF
	}
	msg := r.messages[r.i]
	r.i++
	return msg, nil
}

func (r *messageReader) Close() error {
	return nil
}

type testFileInfo struct {
	name string
	size int64
	time time.Time
	sys  interface{}
}

func (t testFileInfo) Name() string       { return t.name }
func (t testFileInfo) Size() int64        { return t.size }
func (t testFileInfo) Mode() os.FileMode  { return 0 }
func (t testFileInfo) ModTime() time.Time { return t.time }
func (t testFileInfo) IsDir() bool        { return false }
func (t testFileInfo) Sys() interface{}   { return t.sys }
