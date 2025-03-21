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

package memqueue

import "github.com/njcx/libbeat_v8/publisher/queue"

// producer -> broker API

type pushRequest struct {
	event queue.Entry

	// The event's encoded size in bytes if the configured output supports
	// early encoding, 0 otherwise.
	eventSize int

	// The producer that generated this event, or nil if this producer does
	// not require ack callbacks.
	producer *ackProducer

	// The index of the event in this producer only. Used to condense
	// multiple acknowledgments for a producer to a single callback call.
	producerID producerID
	resp       chan queue.EntryID
}

// consumer -> broker API

type getRequest struct {
	entryCount   int         // request entryCount events from the broker
	responseChan chan *batch // channel to send response to
}

type batchDoneMsg struct{}
