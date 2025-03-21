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

// This file contains commonly-used utility functions for testing.

package testutil

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/njcx/libbeat_v8/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	SeedFlag = flag.Int64("seed", 0, "Randomization seed")
)

func SeedPRNG(t *testing.T) {
	seed := *SeedFlag
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	t.Logf("reproduce test with `go test ... -seed %v`", seed)
	rand.New(rand.NewSource(seed))
}

func GenerateEvents(numEvents, fieldsPerLevel, depth int) []beat.Event {
	events := make([]beat.Event, numEvents)
	for i := 0; i < numEvents; i++ {
		event := &beat.Event{Fields: mapstr.M{}}
		generateFields(event, fieldsPerLevel, depth)
		events[i] = *event
	}
	return events
}

func generateFields(event *beat.Event, fieldsPerLevel, depth int) {
	if depth == 0 {
		return
	}

	for j := 1; j <= fieldsPerLevel; j++ {
		var key string
		for d := 1; d <= depth; d++ {
			key += fmt.Sprintf("level%dfield%d", d, j)
			key += "."
		}
		event.Fields.Put(key, "value")
		key = ""
	}

}
