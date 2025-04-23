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
		path := partsToOpenAPIPath(v)
		assert.Equal(t, "/pet/{pet_id}:addPet", path)
	})

	t.Run("with glob pattern", func(t *testing.T) {
		v, err := RunPathPatternLexer("/users/v1/{name=organizations/*/teams/*/members/*}:activate")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v)
		assert.Equal(t, "/users/v1/organizations/{organization}/teams/{team}/members/{member}:activate", path)
	})
}
