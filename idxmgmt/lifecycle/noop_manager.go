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

package lifecycle

import (
	"github.com/njcx/libbeat_v8/beat"
)

type noopSupport struct{}
type noopManager struct{}

// NewNoopSupport creates a noop ILM implementation with ILM support being always
// disabled.  Attempts to install a policy will fail.
func NewNoopSupport(info beat.Info, c bool) (Supporter, error) {
	return (*noopSupport)(nil), nil
}

// Enabled no-op
func (*noopSupport) Enabled() bool { return false }

// Policy no-op
func (*noopSupport) Policy() Policy { return Policy{} }

// Overwrite no-op
func (*noopSupport) Overwrite() bool { return false }

// Manager no-op
func (*noopSupport) Manager(_ ClientHandler) Manager { return (*noopManager)(nil) }

// CheckEnabled no-op
func (*noopManager) CheckEnabled() (bool, error) { return false, nil }

// EnsurePolicy no-op
func (*noopManager) EnsurePolicy(_ bool) (bool, error) { return false, ErrOpNotAvailable }

// Policyname no-op
func (*noopManager) PolicyName() string { return "" }
