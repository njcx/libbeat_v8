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
	_ "github.com/njcx/libbeat_v8/autodiscover/appenders/config" // Register autodiscover appenders
	_ "github.com/njcx/libbeat_v8/autodiscover/providers/jolokia"
	_ "github.com/njcx/libbeat_v8/monitoring/report/elasticsearch" // Register default monitoring reporting
	_ "github.com/njcx/libbeat_v8/processors/actions"              // Register default processors.
	_ "github.com/njcx/libbeat_v8/processors/add_cloud_metadata"
	_ "github.com/njcx/libbeat_v8/processors/add_formatted_index"
	_ "github.com/njcx/libbeat_v8/processors/add_host_metadata"
	_ "github.com/njcx/libbeat_v8/processors/add_id"
	_ "github.com/njcx/libbeat_v8/processors/add_locale"
	_ "github.com/njcx/libbeat_v8/processors/add_observer_metadata"
	_ "github.com/njcx/libbeat_v8/processors/add_process_metadata"
	_ "github.com/njcx/libbeat_v8/processors/communityid"
	_ "github.com/njcx/libbeat_v8/processors/convert"
	_ "github.com/njcx/libbeat_v8/processors/decode_duration"
	_ "github.com/njcx/libbeat_v8/processors/decode_xml"
	_ "github.com/njcx/libbeat_v8/processors/decode_xml_wineventlog"
	_ "github.com/njcx/libbeat_v8/processors/dissect"
	_ "github.com/njcx/libbeat_v8/processors/dns"
	_ "github.com/njcx/libbeat_v8/processors/extract_array"
	_ "github.com/njcx/libbeat_v8/processors/fingerprint"
	_ "github.com/njcx/libbeat_v8/processors/move_fields"
	_ "github.com/njcx/libbeat_v8/processors/ratelimit"
	_ "github.com/njcx/libbeat_v8/processors/registered_domain"
	_ "github.com/njcx/libbeat_v8/processors/script"
	_ "github.com/njcx/libbeat_v8/processors/syslog"
	_ "github.com/njcx/libbeat_v8/processors/translate_ldap_attribute"
	_ "github.com/njcx/libbeat_v8/processors/translate_sid"
	_ "github.com/njcx/libbeat_v8/processors/urldecode"
	_ "github.com/njcx/libbeat_v8/publisher/includes" // Register publisher pipeline modules
)
