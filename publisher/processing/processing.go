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

package processing

import (
	"github.com/njcx/libbeat_v8/beat"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

// SupportFactory creates a new processing Supporter that can be used with
// the publisher pipeline.  The factory gets the global configuration passed,
// in order to configure some shared global event processing.
type SupportFactory func(info beat.Info, log *logp.Logger, cfg *config.C) (Supporter, error)

// Supporter is used to create an event processing pipeline. It is used by the
// publisher pipeline when a client connects to the pipeline. The supporter
// will merge the global and local configurations into a common event
// processor.
// If `drop` is set, then the processor generated must always drop all events.
// A Supporter needs to be closed with `Close()` to release its global resources.
type Supporter interface {
	// Create a running processor interface based on the given config
	Create(cfg beat.ProcessingConfig, drop bool) (beat.Processor, error)
	// Processors returns a list of config strings for the given processor, for debug purposes
	Processors() []string
	// Close the processor supporter
	Close() error
}
