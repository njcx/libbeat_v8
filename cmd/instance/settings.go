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

package instance

import (
	"github.com/spf13/pflag"

	"github.com/njcx/libbeat_v8/cfgfile"
	"github.com/njcx/libbeat_v8/idxmgmt"
	"github.com/njcx/libbeat_v8/idxmgmt/lifecycle"
	"github.com/njcx/libbeat_v8/monitoring/report"
	"github.com/njcx/libbeat_v8/publisher/processing"
)

// Settings contains basic settings for any beat to pass into GenRootCmd
type Settings struct {
	Name            string
	IndexPrefix     string
	Version         string
	HasDashboards   bool
	ElasticLicensed bool
	Monitoring      report.Settings
	RunFlags        *pflag.FlagSet
	ConfigOverrides []cfgfile.ConditionalOverride

	DisableConfigResolver bool

	// load custom index manager. The config object will be the Beats root configuration.
	IndexManagement idxmgmt.SupportFactory
	ILM             lifecycle.SupportFactory

	Processing processing.SupportFactory

	// InputQueueSize is the size for the internal publisher queue in the
	// publisher pipeline. This is only useful when the Beat plans to use
	// beat.DropIfFull PublishMode. Leave as zero for default.
	InputQueueSize int

	// Initialize functions that are called in-order to initialize unique items for the beat.
	Initialize []func()
}
