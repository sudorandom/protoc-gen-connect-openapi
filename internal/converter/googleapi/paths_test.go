package googleapi

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto" // For proto.Bool
	"gopkg.in/yaml.v3"
)

func TestMergeOrAppendParameter(t *testing.T) {
	t.Run("merges new properties into existing parameter without overwriting", func(t *testing.T) {
		existingParams := []*v3.Parameter{
			{ // param1 - will be targeted by newParam1 for merging new properties
				Name:        "param1",
				In:          "query",
				Description: "Original P1 Description", // Should be preserved
				Deprecated:  false,                     // Will be updated by newParam1 because new is true
				Style:       "form",                    // Should be preserved
				Schema: base.CreateSchemaProxy(&base.Schema{
					Title:       "Original P1 Schema Title",       // Should be preserved
					Description: "Original P1 Schema Description", // Should be preserved
					// Type initially nil to test merging Type
					// Format initially empty to test merging Format
					Enum:    []*yaml.Node{{Kind: yaml.ScalarNode, Value: "one"}, {Kind: yaml.ScalarNode, Value: "two"}}, // Should be preserved
					Default: &yaml.Node{Kind: yaml.ScalarNode, Value: "one"},                                          // Should be preserved
					// Items initially nil to test merging Items
				}),
			},
			{ // param2 - will be targeted by newParam2 to test non-overwrite
				Name:            "param2",
				In:              "query",
				Description:     "Original P2 Description",
				Explode:         proto.Bool(true), // Should be preserved
				Required:        proto.Bool(false),// Should be preserved (newParam2 will try to set true)
				Deprecated:      true,             // Should be preserved (newParam2 will try to set false)
				Style:           "spaceDelimited", // Should be preserved
				AllowEmptyValue: true,             // Corrected to bool, Should be preserved
				AllowReserved:   true,             // Should be preserved
				Schema: base.CreateSchemaProxy(&base.Schema{
					Title:       "Original P2 Schema Title",
					Description: "Original P2 Schema Description",
					Type:        []string{"array"},
					Format:      "csv",
					Enum:        []*yaml.Node{{Kind: yaml.ScalarNode, Value: "alpha"}, {Kind: yaml.ScalarNode, Value: "beta"}},
					Default:     &yaml.Node{Kind: yaml.ScalarNode, Value: "beta"},
					Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					})},
				}),
			},
		}

		// Setup param1 for specific merge tests by making some fields explicitly empty/nil
		existingParams[0].Schema.Schema().Type = nil    // For testing Type merge
		existingParams[0].Schema.Schema().Format = ""   // For testing Format merge
		existingParams[0].Schema.Schema().Default = nil // For testing Default merge
		existingParams[0].Schema.Schema().Items = nil   // For testing Items merge

		newParam1 := &v3.Parameter{ // Targets param1
			Name:            "param1",
			In:              "query",
			Description:     "New P1 Description (should be ignored)",
			Required:        proto.Bool(true), // Merged (original is nil)
			Deprecated:      true,             // Merged (original is false)
			AllowEmptyValue: true,             // Merged (original is false for bool type), corrected from proto.Bool(true)
			Style:           "pipeDelimited",  // Should be ignored (original is "form")
			AllowReserved:   true,             // Merged (original is false)
			Schema: base.CreateSchemaProxy(&base.Schema{
				Title:       "New P1 Schema Title (should be ignored)",
				Description: "New P1 Schema Description (should be ignored)",
				Type:        []string{"integer"}, // Merged
				Format:      "int32",             // Merged
				Enum:        []*yaml.Node{{Kind: yaml.ScalarNode, Value: "3"}, {Kind: yaml.ScalarNode, Value: "4"}},   // Should be ignored
				Default:     &yaml.Node{Kind: yaml.ScalarNode, Value: "42", Tag: "!!float"},         // Merged
				Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{ // Merged
					Type: []string{"number"},
				})},
			}),
		}

		newParam2 := &v3.Parameter{ // Targets param2
			Name:            "param2",
			In:              "query",
			Description:     "New P2 Description (should be ignored)",
			Required:        proto.Bool(true),  // Should be ignored (original is non-nil false)
			Deprecated:      false,             // Should be ignored (original is true)
			Explode:         proto.Bool(false), // Should be ignored (original is non-nil true)
			AllowEmptyValue: false,             // Corrected to bool, Should be ignored (original is true)
			Style:           "matrix",          // Should be ignored
			AllowReserved:   false,             // Should be ignored (original is true)
			Schema: base.CreateSchemaProxy(&base.Schema{
				Title:       "New P2 Schema Title (should be ignored)",
				Description: "New P2 Schema Description (should be ignored)",
				Type:        []string{"object"}, // Should be ignored
				Format:      "json",           // Should be ignored
				Enum:        []*yaml.Node{{Kind: yaml.ScalarNode, Value: "x"}, {Kind: yaml.ScalarNode, Value: "y"}},  // Should be ignored
				Default:     &yaml.Node{Kind: yaml.ScalarNode, Value: "y"},              // Should be ignored
				Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{ // Should be ignored
					Type: []string{"boolean"},
				})},
			}),
		}

		updatedParams := mergeOrAppendParameter(existingParams, newParam1)
		updatedParams = mergeOrAppendParameter(updatedParams, newParam2)

		assert.Len(t, updatedParams, 2)

		// --- Assertions for param1 (merged) ---
		p1 := updatedParams[0]
		assert.Equal(t, "param1", p1.Name)
		assert.Equal(t, "query", p1.In)
		assert.Equal(t, "Original P1 Description", p1.Description) // Preserved: Parameter description
		assert.True(t, *p1.Required)                                // Merged: Parameter Required
		assert.True(t, p1.Deprecated)                               // Merged: Parameter Deprecated (original:false, new:true)
		assert.True(t, p1.AllowEmptyValue)                          // Merged: Parameter AllowEmptyValue (original:false, new:true), removed indirection
		assert.Equal(t, "form", p1.Style)                           // Preserved: Parameter Style
		assert.True(t, p1.AllowReserved)                            // Merged: Parameter AllowReserved (original:false, new:true)

		p1Schema := p1.Schema.Schema()
		assert.Equal(t, "Original P1 Schema Title", p1Schema.Title)                 // Preserved: Schema Title
		assert.Equal(t, "Original P1 Schema Description", p1Schema.Description)     // Preserved: Schema Description
		assert.Equal(t, []string{"integer"}, p1Schema.Type)                         // Merged: Schema Type
		assert.Equal(t, "int32", p1Schema.Format)                                  // Merged: Schema Format
		assert.Equal(t, []*yaml.Node{{Kind: yaml.ScalarNode, Value: "one"}, {Kind: yaml.ScalarNode, Value: "two"}}, p1Schema.Enum) // Preserved: Schema Enum
		require.NotNil(t, p1Schema.Default)                                        // Merged: Schema Default
		assert.Equal(t, "42", p1Schema.Default.Value)                               // Merged: Schema Default
		assert.Equal(t, "!!float", p1Schema.Default.Tag)                            // Merged: Schema Default Tag
		require.NotNil(t, p1Schema.Items, "Schema Items should have been merged for p1") // Merged: Schema Items
		assert.Equal(t, []string{"number"}, p1Schema.Items.A.Schema().Type)         // Merged: Schema Items Type

		// --- Assertions for param2 (should have preserved original values) ---
		p2 := updatedParams[1]
		assert.Equal(t, "param2", p2.Name)
		assert.Equal(t, "Original P2 Description", p2.Description)     // Preserved
		assert.True(t, *p2.Explode)                                     // Preserved
		assert.False(t, *p2.Required)                                   // Preserved (was explicitly false)
		assert.True(t, p2.Deprecated)                                   // Preserved (was true)
		assert.Equal(t, "spaceDelimited", p2.Style)                     // Preserved
		assert.True(t, p2.AllowEmptyValue)                              // Preserved, removed indirection
		assert.True(t, p2.AllowReserved)                                // Preserved

		p2Schema := p2.Schema.Schema()
		assert.Equal(t, "Original P2 Schema Title", p2Schema.Title)                 // Preserved
		assert.Equal(t, "Original P2 Schema Description", p2Schema.Description)     // Preserved
		assert.Equal(t, []string{"array"}, p2Schema.Type)                           // Preserved
		assert.Equal(t, "csv", p2Schema.Format)                                     // Preserved
		assert.Equal(t, []*yaml.Node{{Kind: yaml.ScalarNode, Value: "alpha"}, {Kind: yaml.ScalarNode, Value: "beta"}}, p2Schema.Enum) // Preserved
		require.NotNil(t, p2Schema.Default)                                        // Preserved
		assert.Equal(t, "beta", p2Schema.Default.Value)                             // Preserved
		require.NotNil(t, p2Schema.Items, "Schema Items should be present for p2")   // Preserved
		assert.Equal(t, []string{"string"}, p2Schema.Items.A.Schema().Type)         // Preserved
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
