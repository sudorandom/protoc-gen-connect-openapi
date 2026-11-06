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

package googleapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartsToOpenAPIPath(t *testing.T) {
	t.Run("with annotation", func(t *testing.T) {
		v, err := RunPathPatternLexer("/pet/{pet_id}:addPet")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil, nil)
		assert.Equal(t, "/pet/{pet_id}:addPet", path)
	})

	t.Run("with renames", func(t *testing.T) {
		v, err := RunPathPatternLexer("/pet/{pet_id}:addPet")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil, map[string]string{"pet_id": "petId"})
		assert.Equal(t, "/pet/{petId}:addPet", path)
	})

	t.Run("with glob pattern", func(t *testing.T) {
		v, err := RunPathPatternLexer("/users/v1/{name=organizations/*/teams/*/members/*}:activate")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil, nil)
		assert.Equal(t, "/users/v1/organizations/{organization}/teams/{team}/members/{member}:activate", path)
	})

	t.Run("with glob pattern containing literal segment", func(t *testing.T) {
		v, err := RunPathPatternLexer("/users/v1/{name=organizations/*/teams/*/all/members/*}:activate")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil, nil)
		assert.Equal(t, "/users/v1/organizations/{organization}/teams/{team}/all/members/{member}:activate", path)
	})

	t.Run("with resource map", func(t *testing.T) {
		resMap := map[string]string{
			"timeseries": "timeseries",
		}
		v, err := RunPathPatternLexer("/v1/{name=organizations/*/timeseries/*}")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, resMap, nil)
		assert.Equal(t, "/v1/organizations/{organization}/timeseries/{timeseries}", path)
	})

	t.Run("without resource map", func(t *testing.T) {
		v, err := RunPathPatternLexer("/v1/{name=organizations/*/timeseries/*}")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil, nil)
		assert.Equal(t, "/v1/organizations/{organization}/timeseries/{timesery}", path)
	})
}
