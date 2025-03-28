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

package add_formatted_index

import (
	"errors"

	"github.com/njcx/libbeat_v8/common/fmtstr"
)

// configuration for AddFormattedIndex processor.
type config struct {
	Index *fmtstr.TimestampFormatString `config:"index"` // Index formatted string value
}

// Validate ensures that the configuration is valid.
func (c *config) Validate() error {
	// Validate type of ID generator
	if c.Index == nil {
		return errors.New("index field is required")
	}

	return nil
}
