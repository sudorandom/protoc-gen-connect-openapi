// Copyright 2023-2026 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"testing"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleRPCErrorDetailSchemasFieldNames(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		opts     options.Options
		expected string
	}{
		"json names":  {opts: options.Options{}, expected: "fieldViolations"},
		"proto names": {opts: options.Options{WithProtoNames: true}, expected: "field_violations"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			badRequest, ok := GoogleRPCErrorDetailSchemas(test.opts).Get("google.rpc.BadRequest")
			require.True(t, ok)
			_, ok = badRequest.Schema().Properties.Get(test.expected)
			assert.True(t, ok, "missing property %q", test.expected)
		})
	}
}
