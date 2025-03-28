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

package add_host_metadata

import (
	"os"
	"time"

	"github.com/njcx/libbeat_v8/processors/util"
)

// Config for add_host_metadata processor.
type Config struct {
	NetInfoEnabled      bool            `config:"netinfo.enabled"` // Add IP and MAC to event
	CacheTTL            time.Duration   `config:"cache.ttl"`
	ExpireUpdateTimeout time.Duration   `config:"expire_update_timeout"`
	Geo                 *util.GeoConfig `config:"geo"`
	Name                string          `config:"name"`
	ReplaceFields       bool            `config:"replace_fields"` // replace existing host fields with add_host_metadata
}

func defaultConfig() Config {
	// Setting environmental variable ELASTIC_NETINFO:false in Elastic Agent pod will disable the netinfo.enabled option of add_host_metadata processor
	// This will result to events not being enhanced with host.ip and host.mac
	// Related to https://github.com/elastic/integrations/issues/6674
	valueNETINFO, _ := os.LookupEnv("ELASTIC_NETINFO")

	if valueNETINFO == "false" {
		return Config{
			NetInfoEnabled:      false,
			CacheTTL:            5 * time.Minute,
			ExpireUpdateTimeout: time.Second * 10,
			ReplaceFields:       true,
		}
	} else {
		return Config{
			NetInfoEnabled:      true,
			CacheTTL:            5 * time.Minute,
			ExpireUpdateTimeout: time.Second * 10,
			ReplaceFields:       true,
		}
	}
}
