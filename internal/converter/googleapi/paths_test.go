package googleapi

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto" // For proto.Bool
)

func TestMergeOrAppendParameter(t *testing.T) {
	t.Run("merges new properties into existing parameter without overwriting", func(t *testing.T) {
		existingParams := []*v3.Parameter{
			{
				Name:        "search",
				In:          "query",
				Description: "Custom Search Description",
				Schema:      base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			},
			{
				Name:        "type",
				In:          "query",
				Description: "Original Type Description",
				Explode:     proto.Bool(true),
				Schema: base.CreateSchemaProxy(&base.Schema{
					Type: []string{"array"},
					Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					})},
				}),
			},
		}

		newSearchParam := &v3.Parameter{
			Name:        "search",
			In:          "query",
			Description: "Proto search description", // Should be ignored
			Schema:      base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}, Title: "Proto Search Title"}), // Schema title should be added
			Required:    proto.Bool(true),                                                                          // Should be added if existing is nil
		}

		newTypeParam := &v3.Parameter{
			Name:        "type",
			In:          "query",
			Description: "Proto type description", // Should be ignored
			Explode:     proto.Bool(false),      // Should be ignored
			Schema: base.CreateSchemaProxy(&base.Schema{ // Schema should be ignored if more basic
				Type: []string{"array"},
				Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{
					Type: []string{"string"},
				})},
				Description: "Proto schema description", // This could be added to existing schema if logic allows
			}),
		}

		updatedParams := mergeOrAppendParameter(existingParams, newSearchParam)
		updatedParams = mergeOrAppendParameter(updatedParams, newTypeParam)

		assert.Len(t, updatedParams, 2)

		searchParam := updatedParams[0]
		assert.Equal(t, "search", searchParam.Name)
		assert.Equal(t, "query", searchParam.In)
		assert.Equal(t, "Custom Search Description", searchParam.Description) // Original description preserved
		assert.True(t, *searchParam.Required)                                 // New property added
		assert.NotNil(t, searchParam.Schema.Schema())                         // Fixed: Use Schema() method
		assert.Equal(t, "Proto Search Title", searchParam.Schema.Schema().Title) // Fixed: Use Schema() method
		assert.Contains(t, searchParam.Schema.Schema().Type, "string")        // Fixed: Use Schema() method

		typeParam := updatedParams[1]
		assert.Equal(t, "type", typeParam.Name)
		assert.Equal(t, "query", typeParam.In)
		assert.Equal(t, "Original Type Description", typeParam.Description) // Original description preserved
		assert.True(t, *typeParam.Explode)                                  // Original explode preserved
		assert.NotNil(t, typeParam.Schema.Schema())                         // Fixed: Use Schema() method
		assert.Contains(t, typeParam.Schema.Schema().Type, "array")        // Fixed: Use Schema() method
		// assert.Equal(t, "Proto schema description", typeParam.Schema.Schema().Description) // Check if schema description was merged (optional based on merge depth)
	})

	t.Run("appends new parameter if not existing", func(t *testing.T) {
		existingParams := []*v3.Parameter{
			{Name: "id", In: "path"},
		}
		newParam := &v3.Parameter{Name: "limit", In: "query", Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}})}

		updatedParams := mergeOrAppendParameter(existingParams, newParam)
		assert.Len(t, updatedParams, 2)
		assert.Equal(t, "limit", updatedParams[1].Name)
	})
}

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
