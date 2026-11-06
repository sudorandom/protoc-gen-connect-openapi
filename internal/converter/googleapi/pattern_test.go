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

package googleapi_test

import (
	"testing"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunPathPatternLexer(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{{Type: "EOF", Value: ""}}, v)
	})

	t.Run("slash", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-idents", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/pet/store")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "pet"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "store"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-wildcard", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/*/test")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "LITERAL", Value: "*"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "test"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-wildcard", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/test/**")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "test"},
			{Type: "SLASH", Value: "/"},
			{Type: "LITERAL", Value: "**"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-variable", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/pet/{pet_id}")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "pet"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "pet_id"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-annotation", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/pet/{pet_id}:addPet")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "pet"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "pet_id"},
			{Type: "COLON", Value: ":"},
			{Type: "IDENT", Value: "addPet"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("with-subfield", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/v1/messages/{message_id}/{sub.subfield}")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "v1"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "messages"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "message_id"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "sub.subfield"},
			{Type: "EOF"},
		}, v)
	})
	t.Run("with-value-prefix", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/v1/{property_in_path=messages/*}")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "v1"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "property_in_path=messages/*"},
			{Type: "EOF"},
		}, v)
	})

	t.Run("well-known", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/.well-known/jwks.json")
		require.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: ".well-known"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "jwks.json"},
			{Type: "EOF"},
		}, v)
	})
}
