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

package pipeline

import (
	"runtime"
	"sync"
	"testing"

	"github.com/njcx/libbeat_v8/beat"
	"github.com/njcx/libbeat_v8/common/atomic"
	"github.com/njcx/libbeat_v8/publisher/queue"
	"github.com/njcx/libbeat_v8/tests/resources"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestPipelineAcceptsAnyNumberOfClients(t *testing.T) {
	routinesChecker := resources.NewGoroutinesChecker()
	defer routinesChecker.Check(t)

	pipeline := makePipeline(t, Settings{}, makeDiscardQueue())

	defer pipeline.Close()

	n := 66000
	clients := []beat.Client{}
	for i := 0; i < n; i++ {
		c, err := pipeline.ConnectWith(beat.ClientConfig{})
		if err != nil {
			t.Fatalf("Could not connect to pipeline: %s", err)
		}
		clients = append(clients, c)
	}

	for i, c := range clients {
		c.Publish(beat.Event{
			Fields: mapstr.M{
				"count": i,
			},
		})
	}

	// Close the first 105 clients
	nn := 105
	clientsToClose := clients[:n]
	clients = clients[nn:]

	for _, c := range clientsToClose {
		c.Close()
	}

	// Let other goroutines run
	runtime.Gosched()
	runtime.Gosched()

	// Make sure all clients are closed
	for _, c := range clients {
		c.Close()
	}
}

// makeDiscardQueue returns a queue that always discards all events
// the producers are assigned an unique incremental ID, when their
// close method is called, this ID is returned
func makeDiscardQueue() queue.Queue {
	var wg sync.WaitGroup
	producerID := atomic.NewInt(0)

	return &testQueue{
		close: func() error {
			//  Wait for all producers to finish
			wg.Wait()
			return nil
		},
		get: func(count int) (queue.Batch, error) {
			return nil, nil
		},

		producer: func(cfg queue.ProducerConfig) queue.Producer {
			producerID.Inc()

			// count is a counter that increments on every published event
			// it's also the returned Event ID
			count := uint64(0)
			producer := &testProducer{
				publish: func(try bool, event queue.Entry) (queue.EntryID, bool) {
					count++
					return queue.EntryID(count), true
				},
				cancel: func() {
					wg.Done()
				},
			}

			wg.Add(1)
			return producer
		},
	}
}

type testQueue struct {
	close        func() error
	bufferConfig func() queue.BufferConfig
	producer     func(queue.ProducerConfig) queue.Producer
	get          func(sz int) (queue.Batch, error)
}

type testProducer struct {
	publish func(try bool, event queue.Entry) (queue.EntryID, bool)
	cancel  func()
}

func (q *testQueue) Close() error {
	if q.close != nil {
		return q.close()
	}
	return nil
}

func (q *testQueue) Done() <-chan struct{} {
	return nil
}

func (q *testQueue) QueueType() string {
	return "test"
}

func (q *testQueue) BufferConfig() queue.BufferConfig {
	if q.bufferConfig != nil {
		return q.bufferConfig()
	}
	return queue.BufferConfig{}
}

func (q *testQueue) Producer(cfg queue.ProducerConfig) queue.Producer {
	if q.producer != nil {
		return q.producer(cfg)
	}
	return nil
}

func (q *testQueue) Get(sz int) (queue.Batch, error) {
	if q.get != nil {
		return q.get(sz)
	}
	return nil, nil
}

func (p *testProducer) Publish(event queue.Entry) (queue.EntryID, bool) {
	if p.publish != nil {
		return p.publish(false, event)
	}
	return 0, false
}

func (p *testProducer) TryPublish(event queue.Entry) (queue.EntryID, bool) {
	if p.publish != nil {
		return p.publish(true, event)
	}
	return 0, false
}

func (p *testProducer) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}

func makeTestQueue() queue.Queue {
	var mux sync.Mutex
	var wg sync.WaitGroup
	producers := map[queue.Producer]struct{}{}

	return &testQueue{
		close: func() error {
			mux.Lock()
			for producer := range producers {
				producer.Close()
			}
			mux.Unlock()

			wg.Wait()
			return nil
		},
		get: func(count int) (queue.Batch, error) {
			//<-done
			return nil, nil
		},

		producer: func(cfg queue.ProducerConfig) queue.Producer {
			var producer *testProducer
			p := blockingProducer(cfg)
			producer = &testProducer{
				publish: func(try bool, event queue.Entry) (queue.EntryID, bool) {
					if try {
						return p.TryPublish(event)
					}
					return p.Publish(event)
				},
				cancel: func() {
					mux.Lock()
					defer mux.Unlock()
					delete(producers, producer)
					wg.Done()
				},
			}

			mux.Lock()
			defer mux.Unlock()
			producers[producer] = struct{}{}
			wg.Add(1)
			return producer
		},
	}
}

func blockingProducer(_ queue.ProducerConfig) queue.Producer {
	sig := make(chan struct{})
	waiting := atomic.MakeInt(0)

	return &testProducer{
		publish: func(_ bool, _ queue.Entry) (queue.EntryID, bool) {
			waiting.Inc()
			<-sig
			return 0, false
		},

		cancel: func() {
			close(sig)
		},
	}
}
