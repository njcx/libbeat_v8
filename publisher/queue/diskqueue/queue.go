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

package diskqueue

import (
	"errors"
	"fmt"
	"os"

	"github.com/njcx/libbeat_v8/publisher/queue"
	"github.com/elastic/elastic-agent-libs/logp"
)

// The string used to specify this queue in beats configurations.
const QueueType = "disk"

// diskQueue is the internal type representing a disk-based implementation
// of queue.Queue.
type diskQueue struct {
	logger   *logp.Logger
	observer queue.Observer
	settings Settings

	// Metadata related to the segment files.
	segments diskQueueSegments

	// Metadata related to consumer acks / positions of the oldest remaining
	// frame.
	acks *diskQueueACKs

	// The queue's helper loops, each of which is run in its own goroutine.
	readerLoop  *readerLoop
	writerLoop  *writerLoop
	deleterLoop *deleterLoop

	// writing is true if the writer loop is processing a request, false
	// otherwise.
	writing bool

	// If writing is true, then writeRequestSize equals the number of bytes it
	// contained. Used to calculate how much free capacity the queue has left
	// after all scheduled writes have been completed (see canAcceptFrameOfSize).
	writeRequestSize uint64

	// reading is true if the reader loop is processing a request, false
	// otherwise.
	reading bool

	// deleting is true if the deleter loop is processing a request, false
	// otherwise.
	deleting bool

	// The API channel used by diskQueueProducer to write events.
	producerWriteRequestChan chan producerWriteRequest

	// pendingFrames is a list of all incoming data frames that have been
	// accepted by the queue and are waiting to be sent to the writer loop.
	// Segment ids in this list always appear in sorted order, even between
	// requests (that is, a frame added to this list always has segment id
	// at least as high as every previous frame that has ever been added).
	pendingFrames []segmentedFrame

	// blockedProducers is a list of all producer write requests that are
	// waiting for free space in the queue.
	blockedProducers []producerWriteRequest

	// The channel to signal our goroutines to shut down, used by
	// (*diskQueue).Close.
	close chan struct{}

	// The channel to report that shutdown is finished, used by
	// (*diskQueue).Done.
	done chan struct{}
}

// FactoryForSettings is a simple wrapper around NewQueue so a concrete
// Settings object can be wrapped in a queue-agnostic interface for
// later use by the pipeline.
func FactoryForSettings(settings Settings) queue.QueueFactory {
	return func(
		logger *logp.Logger,
		observer queue.Observer,
		inputQueueSize int,
		encoderFactory queue.EncoderFactory,
	) (queue.Queue, error) {
		return NewQueue(logger, observer, settings, encoderFactory)
	}
}

// NewQueue returns a disk-based queue configured with the given logger
// and settings, creating it if it doesn't exist.
func NewQueue(
	logger *logp.Logger,
	observer queue.Observer,
	settings Settings,
	encoderFactory queue.EncoderFactory,
) (*diskQueue, error) {
	logger = logger.Named("diskqueue")
	logger.Debugf(
		"Initializing disk queue at path %v", settings.directoryPath())
	if observer == nil {
		observer = queue.NewQueueObserver(nil)
	}

	if settings.MaxBufferSize > 0 &&
		settings.MaxBufferSize < settings.MaxSegmentSize*2 {
		return nil, fmt.Errorf(
			"disk queue buffer size (%v) must be at least "+
				"twice the segment size (%v)",
			settings.MaxBufferSize, settings.MaxSegmentSize)
	}
	observer.MaxBytes(int(settings.MaxBufferSize))

	// Create the given directory path if it doesn't exist.
	err := os.MkdirAll(settings.directoryPath(), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("couldn't create disk queue directory: %w", err)
	}

	// Load the previous queue position, if any.
	nextReadPosition, err := queuePositionFromPath(settings.stateFilePath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// Errors reading / writing the position are non-fatal -- we just log a
		// warning and fall back on the oldest existing segment, if any.
		logger.Warnf("Couldn't load most recent queue position: %v", err)
	}
	if nextReadPosition.frameIndex == 0 {
		// If the previous state was written by an older version, it may lack
		// the frameIndex field. In this case we reset the read offset within
		// the segment, which may cause one-time retransmission of some events
		// from a previous version, but ensures that our metrics are consistent.
		// In the more common case that frameIndex is 0 because this segment
		// simply hasn't been read yet, setting byteIndex to 0 is a no-op.
		nextReadPosition.byteIndex = 0
	}
	positionFile, err := os.OpenFile(
		settings.stateFilePath(), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		// This is not the _worst_ error: we could try operating even without a
		// position file. But it indicates a problem with the queue permissions on
		// disk, which keeps us from tracking our position within the segment files
		// and could also prevent us from creating new ones, so we treat this as a
		// fatal error on startup rather than quietly providing degraded
		// performance.
		return nil, fmt.Errorf("couldn't write to state file: %w", err)
	}

	// Index any existing data segments to be placed in segments.reading.
	initialSegments, err :=
		scanExistingSegments(logger, settings.directoryPath())
	if err != nil {
		return nil, err
	}
	var nextSegmentID segmentID
	if len(initialSegments) > 0 {
		// Initialize nextSegmentID to the first ID after the existing segments.
		lastID := initialSegments[len(initialSegments)-1].id
		nextSegmentID = lastID + 1
	}
	// Check the initial contents to report to the metrics observer.
	initialEventCount := 0
	initialByteCount := 0
	for _, segment := range initialSegments {
		initialEventCount += int(segment.frameCount)
		// Event metrics for the queue observer don't include segment headser size
		initialByteCount += int(segment.byteCount - segment.headerSize())
	}
	observer.Restore(initialEventCount, initialByteCount)

	// If any of the initial segments are older than the current queue
	// position, move them directly to the acked list where they can be
	// deleted.
	ackedSegments := []*queueSegment{}
	readSegmentID := nextReadPosition.segmentID
	for len(initialSegments) > 0 && initialSegments[0].id < readSegmentID {
		ackedSegments = append(ackedSegments, initialSegments[0])
		initialSegments = initialSegments[1:]
	}

	// If the queue position is older than all existing segments, advance
	// it to the beginning of the first one.
	if len(initialSegments) > 0 && readSegmentID < initialSegments[0].id {
		nextReadPosition = queuePosition{segmentID: initialSegments[0].id}
	}

	// Count just the active events to report in the log
	activeFrameCount := 0
	for _, segment := range initialSegments {
		activeFrameCount += int(segment.frameCount)
	}
	activeFrameCount -= int(nextReadPosition.frameIndex)
	logger.Infof("Found %v queued events consuming %v bytes, %v events still pending", initialEventCount, initialByteCount, activeFrameCount)

	var encoder queue.Encoder
	if encoderFactory != nil {
		encoder = encoderFactory()
	}

	queue := &diskQueue{
		logger:   logger,
		observer: observer,
		settings: settings,

		segments: diskQueueSegments{
			reading:          initialSegments,
			acked:            ackedSegments,
			nextID:           nextSegmentID,
			nextReadPosition: nextReadPosition.byteIndex,
		},

		acks: newDiskQueueACKs(logger, nextReadPosition, positionFile),

		readerLoop:  newReaderLoop(settings, encoder),
		writerLoop:  newWriterLoop(logger, settings),
		deleterLoop: newDeleterLoop(settings),

		producerWriteRequestChan: make(chan producerWriteRequest),

		close: make(chan struct{}),
		done:  make(chan struct{}),
	}

	// Start the goroutines and return the queue!
	go queue.readerLoop.run()
	go queue.writerLoop.run()
	go queue.deleterLoop.run()
	go queue.run()

	return queue, nil
}

//
// diskQueue implementation of the queue.Queue interface
//

func (dq *diskQueue) Close() error {
	// Closing the done channel signals to the core loop that it should
	// shut down the other helper goroutines and wrap everything up.
	close(dq.close)

	return nil
}

func (dq *diskQueue) Done() <-chan struct{} {
	return dq.done
}

func (dq *diskQueue) QueueType() string {
	return QueueType
}

func (dq *diskQueue) BufferConfig() queue.BufferConfig {
	return queue.BufferConfig{MaxEvents: 0}
}

func (dq *diskQueue) Producer(cfg queue.ProducerConfig) queue.Producer {
	return &diskQueueProducer{
		queue:   dq,
		config:  cfg,
		encoder: newEventEncoder(SerializationCBOR),
		done:    make(chan struct{}),
	}
}
