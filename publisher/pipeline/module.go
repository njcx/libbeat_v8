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
	"flag"

	"go.elastic.co/apm/v2"

	"github.com/njcx/libbeat_v8/beat"
	"github.com/njcx/libbeat_v8/outputs"
	"github.com/njcx/libbeat_v8/publisher/processing"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/monitoring"
)

// Global pipeline module for loading the main pipeline from a configuration object

// command line flags
var publishDisabled = false

const defaultQueueType = "mem"

// Monitors configures visibility for observing state and progress of the
// pipeline.
type Monitors struct {
	Metrics   *monitoring.Registry
	Telemetry *monitoring.Registry
	Logger    *logp.Logger
	Tracer    *apm.Tracer
}

// OutputFactory is used by the publisher pipeline to create an output instance.
// If the group returned can be empty. The pipeline will accept events, but
// eventually block.
type outputFactory func(outputs.Observer) (string, outputs.Group, error)

func init() {
	flag.BoolVar(&publishDisabled, "N", false, "Disable actual publishing for testing")
}

// Load uses a Config object to create a new complete Pipeline instance with
// configured queue and outputs. This is a non-blocking operation, and outputs should connect lazily.
func Load(
	beatInfo beat.Info,
	monitors Monitors,
	config Config,
	processors processing.Supporter,
	makeOutput func(outputs.Observer) (string, outputs.Group, error),
) (*Pipeline, error) {

	settings := Settings{
		WaitClose:     0,
		WaitCloseMode: NoWaitOnClose,
		Processors:    processors,
	}

	return LoadWithSettings(beatInfo, monitors, config, makeOutput, settings)
}

// LoadWithSettings is the same as Load, but it exposes a Settings object that includes processors and WaitClose behavior
func LoadWithSettings(
	beatInfo beat.Info,
	monitors Monitors,
	config Config,
	makeOutput func(outputs.Observer) (string, outputs.Group, error),
	settings Settings,
) (*Pipeline, error) {
	log := monitors.Logger
	if log == nil {
		log = logp.L()
	}

	if publishDisabled {
		log.Info("Dry run mode. All output types except the file based one are disabled.")
	}

	name := beatInfo.Name

	out, err := loadOutput(monitors, makeOutput)
	if err != nil {
		return nil, err
	}

	p, err := New(beatInfo, monitors, config.Queue, out, settings)
	if err != nil {
		return nil, err
	}

	log.Infof("Beat name: %s", name)
	return p, err
}

func loadOutput(
	monitors Monitors,
	makeOutput outputFactory,
) (outputs.Group, error) {
	if publishDisabled {
		return outputs.Group{}, nil
	}

	if makeOutput == nil {
		return outputs.Group{}, nil
	}

	var (
		metrics  *monitoring.Registry
		outStats outputs.Observer
	)
	if monitors.Metrics != nil {
		metrics = monitors.Metrics.GetRegistry("output")
		if metrics != nil {
			err := metrics.Clear()
			if err != nil {
				return outputs.Group{}, err
			}

		} else {
			metrics = monitors.Metrics.NewRegistry("output")
		}
		outStats = outputs.NewStats(metrics)
	}

	outName, out, err := makeOutput(outStats)
	if err != nil {
		return outputs.Fail(err)
	}

	if metrics != nil {
		monitoring.NewString(metrics, "type").Set(outName)
	}
	if monitors.Telemetry != nil {
		telemetry := monitors.Telemetry.GetRegistry("output")
		if telemetry != nil {
			err := telemetry.Clear()
			if err != nil {
				return outputs.Group{}, err
			}
		} else {
			telemetry = monitors.Telemetry.NewRegistry("output")
		}
		monitoring.NewString(telemetry, "name").Set(outName)
		monitoring.NewInt(telemetry, "batch_size").Set(int64(out.BatchSize))
		monitoring.NewInt(telemetry, "clients").Set(int64(len(out.Clients)))
	}

	return out, nil
}
