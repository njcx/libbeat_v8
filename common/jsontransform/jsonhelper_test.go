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

package jsontransform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/njcx/libbeat_v8/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestWriteJSONKeys(t *testing.T) {
	now := time.Now()
	now = now.Round(time.Second)

	eventTimestamp := time.Date(2020, 01, 01, 01, 01, 00, 0, time.UTC)
	eventMetadata := mapstr.M{
		"foo": "bar",
		"baz": mapstr.M{
			"qux": 17,
		},
	}
	eventFields := mapstr.M{
		"top_a": 23,
		"top_b": mapstr.M{
			"inner_c": "see",
			"inner_d": "dee",
		},
	}

	tests := map[string]struct {
		keys              map[string]interface{}
		expandKeys        bool
		overwriteKeys     bool
		expectedMetadata  mapstr.M
		expectedTimestamp time.Time
		expectedFields    mapstr.M
		addErrorKeys      bool
	}{
		"overwrite_true": {
			overwriteKeys: true,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(time.RFC3339),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedMetadata: mapstr.M{
				"foo": "NEW_bar",
				"baz": mapstr.M{
					"qux":   "NEW_qux",
					"durrr": "COMPLETELY_NEW",
				},
			},
			expectedTimestamp: now,
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"overwrite_true_ISO8601": {
			overwriteKeys: true,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(iso8601),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedMetadata: mapstr.M{
				"foo": "NEW_bar",
				"baz": mapstr.M{
					"qux":   "NEW_qux",
					"durrr": "COMPLETELY_NEW",
				},
			},
			expectedTimestamp: now,
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"overwrite_false": {
			overwriteKeys: false,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(time.RFC3339),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedMetadata:  eventMetadata.Clone(),
			expectedTimestamp: eventTimestamp,
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"expand_true": {
			expandKeys:    true,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			expectedMetadata:  eventMetadata.Clone(),
			expectedTimestamp: eventTimestamp,
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": mapstr.M{
						"inner_e": "COMPLETELY_NEW_e",
					},
				},
			},
		},
		"expand_false": {
			expandKeys:    false,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			expectedMetadata:  eventMetadata.Clone(),
			expectedTimestamp: eventTimestamp,
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c":         "see",
					"inner_d":         "dee",
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
		},
		// This benchmark makes sure that when an error is found in the event, the proper fields are defined and measured
		"error_case": {
			expandKeys:    false,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
				"@timestamp": map[string]interface{}{"when": "now", "another": "yesterday"},
			},
			expectedMetadata:  eventMetadata.Clone(),
			expectedTimestamp: eventTimestamp,
			expectedFields: mapstr.M{
				"error": mapstr.M{
					"message": "@timestamp not overwritten (not string)",
					"type":    "json",
				},
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c":         "see",
					"inner_d":         "dee",
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			addErrorKeys: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			event := &beat.Event{
				Timestamp: eventTimestamp,
				Meta:      eventMetadata.Clone(),
				Fields:    eventFields.Clone(),
			}

			WriteJSONKeys(event, test.keys, test.expandKeys, test.overwriteKeys, test.addErrorKeys)
			require.Equal(t, test.expectedMetadata, event.Meta)
			require.Equal(t, test.expectedTimestamp.UnixNano(), event.Timestamp.UnixNano())
			require.Equal(t, test.expectedFields, event.Fields)
		})
	}
}

func BenchmarkWriteJSONKeys(b *testing.B) {
	now := time.Now()
	now = now.Round(time.Second)

	eventTimestamp := time.Date(2020, 01, 01, 01, 01, 00, 0, time.UTC)
	eventMetadata := mapstr.M{
		"foo": "bar",
		"baz": mapstr.M{
			"qux": 17,
		},
	}
	eventFields := mapstr.M{
		"top_a": 23,
		"top_b": mapstr.M{
			"inner_c": "see",
			"inner_d": "dee",
		},
	}

	tests := map[string]struct {
		keys           map[string]interface{}
		expandKeys     bool
		overwriteKeys  bool
		expectedFields mapstr.M
		addErrorKeys   bool
	}{
		"overwrite_true": {
			overwriteKeys: true,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(time.RFC3339),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"overwrite_true_ISO8601": {
			overwriteKeys: true,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(iso8601),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"overwrite_false": {
			overwriteKeys: false,
			keys: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"foo": "NEW_bar",
					"baz": map[string]interface{}{
						"qux":   "NEW_qux",
						"durrr": "COMPLETELY_NEW",
					},
				},
				"@timestamp": now.Format(time.RFC3339),
				"top_b": map[string]interface{}{
					"inner_d": "NEW_dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": "dee",
					"inner_e": "COMPLETELY_NEW_e",
				},
				"top_c": "COMPLETELY_NEW_c",
			},
		},
		"expand_true": {
			expandKeys:    true,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c": "see",
					"inner_d": mapstr.M{
						"inner_e": "COMPLETELY_NEW_e",
					},
				},
			},
		},
		"expand_false": {
			expandKeys:    false,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			expectedFields: mapstr.M{
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c":         "see",
					"inner_d":         "dee",
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
		},
		// This benchmark makes sure that when an error is found in the event, the proper fields are defined and measured
		"error_case": {
			expandKeys:    false,
			overwriteKeys: true,
			keys: map[string]interface{}{
				"top_b": map[string]interface{}{
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
				"@timestamp": "invalid string",
			},
			expectedFields: mapstr.M{
				"error": mapstr.M{
					"message": "@timestamp not overwritten (parse error on invalid string)",
					"type":    "json",
				},
				"top_a": 23,
				"top_b": mapstr.M{
					"inner_c":         "see",
					"inner_d":         "dee",
					"inner_d.inner_e": "COMPLETELY_NEW_e",
				},
			},
			addErrorKeys: true,
		},
	}

	for name, test := range tests {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				event := &beat.Event{
					Timestamp: eventTimestamp,
					Meta:      eventMetadata.Clone(),
					Fields:    eventFields.Clone(),
				}
				// The WriteJSONKeys will override the keys, so we need to clone it.
				keys := clone(test.keys)
				b.StartTimer()
				WriteJSONKeys(event, keys, test.expandKeys, test.overwriteKeys, test.addErrorKeys)
				require.Equal(b, test.expectedFields, event.Fields)
			}
		})
	}
}

func clone(a map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range a {
		newMap[k] = v
	}
	return newMap
}
