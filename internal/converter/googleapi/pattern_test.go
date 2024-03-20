package googleapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
)

func TestRunPathPatternLexer(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{{Type: "EOF", Value: ""}}, v)
	})

	t.Run("slash", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "EOF", Value: ""},
		}, v)
	})

	t.Run("with-idents", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/pet/store")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "pet"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "store"},
			{Type: "EOF", Value: ""},
		}, v)
	})

	t.Run("with-wildcard", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/*/test")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "LITERAL", Value: "*"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "test"},
			{Type: "EOF", Value: ""},
		}, v)
	})

	t.Run("with-wildcard", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/test/**")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "test"},
			{Type: "SLASH", Value: "/"},
			{Type: "LITERAL", Value: "**"},
			{Type: "EOF", Value: ""},
		}, v)
	})

	t.Run("with-variable", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/pet/{pet_id}")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "pet"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "pet_id"},
			{Type: "EOF", Value: ""},
		}, v)
	})

	t.Run("with-subfield", func(t *testing.T) {
		v, err := googleapi.RunPathPatternLexer("/v1/messages/{message_id}/{sub.subfield}")
		assert.NoError(t, err)
		assert.Equal(t, []googleapi.Token{
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "v1"},
			{Type: "SLASH", Value: "/"},
			{Type: "IDENT", Value: "messages"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "message_id"},
			{Type: "SLASH", Value: "/"},
			{Type: "VARIABLE", Value: "sub.subfield"},
			{Type: "EOF", Value: ""},
		}, v)
	})
}
