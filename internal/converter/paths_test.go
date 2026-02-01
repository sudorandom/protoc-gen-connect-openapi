package converter

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMergeParameters(t *testing.T) {
	tests := []struct {
		name     string
		existing []*v3.Parameter
		new      []*v3.Parameter
		verify   func(t *testing.T, result []*v3.Parameter)
	}{
		{
			name:     "returns new parameters when existing is empty",
			existing: nil,
			new: []*v3.Parameter{
				{Name: "id", In: "path", Description: "Resource ID"},
				{Name: "filter", In: "query", Description: "Filter results"},
			},
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 2)
				assert.Equal(t, "id", result[0].Name)
				assert.Equal(t, "filter", result[1].Name)
			},
		},
		{
			name: "returns existing parameters when new is empty",
			existing: []*v3.Parameter{
				{Name: "id", In: "path", Description: "Resource ID"},
			},
			new: nil,
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 1)
				assert.Equal(t, "id", result[0].Name)
			},
		},
		{
			name: "merges parameters with same name and location",
			existing: []*v3.Parameter{
				{
					Name:        "id",
					In:          "query",
					Description: "Original description",
					Required:    proto.Bool(false),
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			new: []*v3.Parameter{
				{
					Name:     "id",
					In:       "query",
					Required: proto.Bool(true),
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:   []string{"string"},
						Format: "uuid",
					}),
				},
			},
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 1)
				assert.Equal(t, "Original description", result[0].Description, "should preserve original description")
				assert.False(t, *result[0].Required, "should preserve original Required value when set")
				assert.Equal(t, "uuid", result[0].Schema.Schema().Format, "should merge format from new parameter")
			},
		},
		{
			name: "appends parameters with different names",
			existing: []*v3.Parameter{
				{Name: "id", In: "path"},
			},
			new: []*v3.Parameter{
				{Name: "filter", In: "query"},
				{Name: "limit", In: "query"},
			},
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 3)
				assert.Equal(t, "id", result[0].Name)
				assert.Equal(t, "filter", result[1].Name)
				assert.Equal(t, "limit", result[2].Name)
			},
		},
		{
			name: "appends parameters with same name but different location",
			existing: []*v3.Parameter{
				{Name: "id", In: "path"},
			},
			new: []*v3.Parameter{
				{Name: "id", In: "query"},
			},
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 2, "should append parameter with same name but different location")
				assert.Equal(t, "path", result[0].In)
				assert.Equal(t, "query", result[1].In)
			},
		},
		{
			name: "merges multiple parameters with various scenarios",
			existing: []*v3.Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "Path ID",
				},
				{
					Name:        "filter",
					In:          "query",
					Description: "Original filter",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"string"},
					}),
				},
			},
			new: []*v3.Parameter{
				{
					Name: "filter",
					In:   "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:   []string{"string"},
						Format: "regex",
					}),
				},
				{
					Name: "limit",
					In:   "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"integer"},
					}),
				},
			},
			verify: func(t *testing.T, result []*v3.Parameter) {
				assert.Len(t, result, 3)
				assert.Equal(t, "id", result[0].Name)
				assert.Equal(t, "filter", result[1].Name)
				assert.Equal(t, "Original filter", result[1].Description, "should preserve description")
				assert.Equal(t, "regex", result[1].Schema.Schema().Format, "should merge format")
				assert.Equal(t, "limit", result[2].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeParameters(tt.existing, tt.new)
			tt.verify(t, result)
		})
	}
}

func TestMergeOperation(t *testing.T) {
	t.Run("merges parameters", func(t *testing.T) {
		existing := &v3.Operation{
			Parameters: []*v3.Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "Original ID",
					Required:    proto.Bool(true),
				},
				{
					Name: "filter",
					In:   "query",
				},
			},
		}
		newOp := &v3.Operation{
			Parameters: []*v3.Parameter{
				{
					Name: "id",
					In:   "path",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type:   []string{"string"},
						Format: "uuid",
					}),
				},
				{
					Name: "limit",
					In:   "query",
					Schema: base.CreateSchemaProxy(&base.Schema{
						Type: []string{"integer"},
					}),
				},
			},
		}

		mergeOperation(&existing, newOp)
		assert.Len(t, existing.Parameters, 3)

		// Check that id parameter was merged
		idParam := existing.Parameters[0]
		assert.Equal(t, "id", idParam.Name)
		assert.Equal(t, "Original ID", idParam.Description, "should preserve description")
		assert.True(t, *idParam.Required, "should preserve required")
		assert.Equal(t, "uuid", idParam.Schema.Schema().Format, "should merge format")

		// Check that filter parameter is unchanged
		assert.Equal(t, "filter", existing.Parameters[1].Name)

		// Check that limit parameter was appended
		assert.Equal(t, "limit", existing.Parameters[2].Name)
	})
}
