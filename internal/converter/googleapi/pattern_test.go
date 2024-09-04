package googleapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
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
