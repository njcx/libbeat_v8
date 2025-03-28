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

import (
	"github.com/njcx/libbeat_v8/publisher/queue"
	"github.com/elastic/elastic-agent-libs/logp"
)

type forgetfulProducer struct {
	broker    *broker
	openState openState
}

type ackProducer struct {
	broker        *broker
	producedCount uint64
	state         produceState
	openState     openState
}

type openState struct {
	log          *logp.Logger
	done         chan struct{}
	queueClosing <-chan struct{}
	events       chan pushRequest
	encoder      queue.Encoder
}

// producerID stores the order of events within a single producer, so multiple
// event acknowledgement callbacks can be coalesced into a single call.
// It is defined as an explicit type to reduce cross-confusion with the id
// of an event in the queue itself, which is a queue.EntryID.
type producerID uint64

type produceState struct {
	cb      ackHandler
	lastACK producerID
}

type ackHandler func(count int)

func newProducer(b *broker, cb ackHandler, encoder queue.Encoder) queue.Producer {
	openState := openState{
		log:          b.logger,
		done:         make(chan struct{}),
		queueClosing: b.closingChan,
		events:       b.pushChan,
		encoder:      encoder,
	}

	if cb != nil {
		p := &ackProducer{broker: b, openState: openState}
		p.state.cb = cb
		return p
	}
	return &forgetfulProducer{broker: b, openState: openState}
}

func (p *forgetfulProducer) makePushRequest(event queue.Entry) pushRequest {
	resp := make(chan queue.EntryID, 1)
	return pushRequest{
		event: event,
		resp:  resp}
}

func (p *forgetfulProducer) Publish(event queue.Entry) (queue.EntryID, bool) {
	return p.openState.publish(p.makePushRequest(event))
}

func (p *forgetfulProducer) TryPublish(event queue.Entry) (queue.EntryID, bool) {
	return p.openState.tryPublish(p.makePushRequest(event))
}

func (p *forgetfulProducer) Close() {
	p.openState.Close()
}

func (p *ackProducer) makePushRequest(event queue.Entry) pushRequest {
	resp := make(chan queue.EntryID, 1)
	return pushRequest{
		event:    event,
		producer: p,
		// We add 1 to the id so the default lastACK of 0 is a
		// valid initial state and 1 is the first real id.
		producerID: producerID(p.producedCount + 1),
		resp:       resp}
}

func (p *ackProducer) Publish(event queue.Entry) (queue.EntryID, bool) {
	id, published := p.openState.publish(p.makePushRequest(event))
	if published {
		p.producedCount++
	}
	return id, published
}

func (p *ackProducer) TryPublish(event queue.Entry) (queue.EntryID, bool) {
	id, published := p.openState.tryPublish(p.makePushRequest(event))
	if published {
		p.producedCount++
	}
	return id, published
}

func (p *ackProducer) Close() {
	p.openState.Close()
}

func (st *openState) Close() {
	close(st.done)
}

func (st *openState) publish(req pushRequest) (queue.EntryID, bool) {
	// If we were given an encoder callback for incoming events, apply it before
	// sending the entry to the queue.
	if st.encoder != nil {
		req.event, req.eventSize = st.encoder.EncodeEntry(req.event)
	}
	select {
	case st.events <- req:
		// The events channel is buffered, which means we may successfully
		// write to it even if the queue is shutting down. To avoid blocking
		// forever during shutdown, we also have to wait on the queue's
		// shutdown channel.
		select {
		case resp := <-req.resp:
			return resp, true
		case <-st.queueClosing:
			st.events = nil
			return 0, false
		}
	case <-st.done:
		st.events = nil
		return 0, false
	case <-st.queueClosing:
		st.events = nil
		return 0, false
	}
}

func (st *openState) tryPublish(req pushRequest) (queue.EntryID, bool) {
	// If we were given an encoder callback for incoming events, apply it before
	// sending the entry to the queue.
	if st.encoder != nil {
		req.event, req.eventSize = st.encoder.EncodeEntry(req.event)
	}
	select {
	case st.events <- req:
		// The events channel is buffered, which means we may successfully
		// write to it even if the queue is shutting down. To avoid blocking
		// forever during shutdown, we also have to wait on the queue's
		// shutdown channel.
		select {
		case resp := <-req.resp:
			return resp, true
		case <-st.queueClosing:
			st.events = nil
			return 0, false
		}
	case <-st.done:
		st.events = nil
		return 0, false
	default:
		st.log.Debugf("Dropping event, queue is blocked")
		return 0, false
	}
}
