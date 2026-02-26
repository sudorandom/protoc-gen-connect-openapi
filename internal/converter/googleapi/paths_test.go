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
		path := partsToOpenAPIPath(v, nil)
		assert.Equal(t, "/pet/{pet_id}:addPet", path)
	})

	t.Run("with glob pattern", func(t *testing.T) {
		v, err := RunPathPatternLexer("/users/v1/{name=organizations/*/teams/*/members/*}:activate")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil)
		assert.Equal(t, "/users/v1/organizations/{organization}/teams/{team}/members/{member}:activate", path)
	})

	t.Run("with glob pattern containing literal segment", func(t *testing.T) {
		v, err := RunPathPatternLexer("/users/v1/{name=organizations/*/teams/*/all/members/*}:activate")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil)
		assert.Equal(t, "/users/v1/organizations/{organization}/teams/{team}/all/members/{member}:activate", path)
	})

	t.Run("with renames applies camelCase to simple variable", func(t *testing.T) {
		v, err := RunPathPatternLexer("/v1/resources/{resource_id}")
		require.NoError(t, err)
		renames := map[string]string{"resource_id": "resourceId"}
		path := partsToOpenAPIPath(v, renames)
		assert.Equal(t, "/v1/resources/{resourceId}", path)
	})

	t.Run("with renames applies camelCase to multiple variables", func(t *testing.T) {
		v, err := RunPathPatternLexer("/v1/{account_id}/items/{item_id}")
		require.NoError(t, err)
		renames := map[string]string{"account_id": "accountId", "item_id": "itemId"}
		path := partsToOpenAPIPath(v, renames)
		assert.Equal(t, "/v1/{accountId}/items/{itemId}", path)
	})

	t.Run("with renames applies camelCase to glob pattern name", func(t *testing.T) {
		v, err := RunPathPatternLexer("/v1/{resource_name=**}")
		require.NoError(t, err)
		renames := map[string]string{"resource_name": "resourceName"}
		path := partsToOpenAPIPath(v, renames)
		assert.Equal(t, "/v1/{resourceName}", path)
	})

	t.Run("without renames keeps original variable names", func(t *testing.T) {
		v, err := RunPathPatternLexer("/v1/resources/{resource_id}")
		require.NoError(t, err)
		path := partsToOpenAPIPath(v, nil)
		assert.Equal(t, "/v1/resources/{resource_id}", path)
	})
}
