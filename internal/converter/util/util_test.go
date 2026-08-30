package util

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	yamlv3 "go.yaml.in/yaml/v3"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/proto"
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
					Default: &yaml.Node{Kind: yaml.ScalarNode, Value: "one"},                                            // Should be preserved
					// Items initially nil to test merging Items
				}),
			},
			{ // param2 - will be targeted by newParam2 to test non-overwrite
				Name:            "param2",
				In:              "query",
				Description:     "Original P2 Description",
				Explode:         proto.Bool(true),  // Should be preserved
				Required:        proto.Bool(false), // Should be preserved (newParam2 will try to set true)
				Deprecated:      true,              // Should be preserved (newParam2 will try to set false)
				Style:           "spaceDelimited",  // Should be preserved
				AllowEmptyValue: true,              // Corrected to bool, Should be preserved
				AllowReserved:   true,              // Should be preserved
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
				Type:        []string{"integer"},                                                                    // Merged
				Format:      "int32",                                                                                // Merged
				Enum:        []*yaml.Node{{Kind: yaml.ScalarNode, Value: "3"}, {Kind: yaml.ScalarNode, Value: "4"}}, // Should be ignored
				Default:     &yaml.Node{Kind: yaml.ScalarNode, Value: "42", Tag: "!!float"},                         // Merged
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
				Type:        []string{"object"},                                                                     // Should be ignored
				Format:      "json",                                                                                 // Should be ignored
				Enum:        []*yaml.Node{{Kind: yaml.ScalarNode, Value: "x"}, {Kind: yaml.ScalarNode, Value: "y"}}, // Should be ignored
				Default:     &yaml.Node{Kind: yaml.ScalarNode, Value: "y"},                                          // Should be ignored
				Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{ // Should be ignored
					Type: []string{"boolean"},
				})},
			}),
		}

		updatedParams := MergeOrAppendParameter(existingParams, newParam1)
		updatedParams = MergeOrAppendParameter(updatedParams, newParam2)

		assert.Len(t, updatedParams, 2)

		// --- Assertions for param1 (merged) ---
		p1 := updatedParams[0]
		assert.Equal(t, "param1", p1.Name)
		assert.Equal(t, "query", p1.In)
		assert.Equal(t, "Original P1 Description", p1.Description) // Preserved: Parameter description
		assert.True(t, *p1.Required)                               // Merged: Parameter Required
		assert.True(t, p1.Deprecated)                              // Merged: Parameter Deprecated (original:false, new:true)
		assert.True(t, p1.AllowEmptyValue)                         // Merged: Parameter AllowEmptyValue (original:false, new:true), removed indirection
		assert.Equal(t, "form", p1.Style)                          // Preserved: Parameter Style
		assert.True(t, p1.AllowReserved)                           // Merged: Parameter AllowReserved (original:false, new:true)

		p1Schema := p1.Schema.Schema()
		assert.Equal(t, "Original P1 Schema Title", p1Schema.Title)                                                                // Preserved: Schema Title
		assert.Equal(t, "Original P1 Schema Description", p1Schema.Description)                                                    // Preserved: Schema Description
		assert.Equal(t, []string{"integer"}, p1Schema.Type)                                                                        // Merged: Schema Type
		assert.Equal(t, "int32", p1Schema.Format)                                                                                  // Merged: Schema Format
		assert.Equal(t, []*yaml.Node{{Kind: yaml.ScalarNode, Value: "one"}, {Kind: yaml.ScalarNode, Value: "two"}}, p1Schema.Enum) // Preserved: Schema Enum
		require.NotNil(t, p1Schema.Default)                                                                                        // Merged: Schema Default
		assert.Equal(t, "42", p1Schema.Default.Value)                                                                              // Merged: Schema Default
		assert.Equal(t, "!!float", p1Schema.Default.Tag)                                                                           // Merged: Schema Default Tag
		require.NotNil(t, p1Schema.Items, "Schema Items should have been merged for p1")                                           // Merged: Schema Items
		assert.Equal(t, []string{"number"}, p1Schema.Items.A.Schema().Type)                                                        // Merged: Schema Items Type

		// --- Assertions for param2 (should have preserved original values) ---
		p2 := updatedParams[1]
		assert.Equal(t, "param2", p2.Name)
		assert.Equal(t, "Original P2 Description", p2.Description) // Preserved
		assert.True(t, *p2.Explode)                                // Preserved
		assert.False(t, *p2.Required)                              // Preserved (was explicitly false)
		assert.True(t, p2.Deprecated)                              // Preserved (was true)
		assert.Equal(t, "spaceDelimited", p2.Style)                // Preserved
		assert.True(t, p2.AllowEmptyValue)                         // Preserved, removed indirection
		assert.True(t, p2.AllowReserved)                           // Preserved

		p2Schema := p2.Schema.Schema()
		assert.Equal(t, "Original P2 Schema Title", p2Schema.Title)                                                                   // Preserved
		assert.Equal(t, "Original P2 Schema Description", p2Schema.Description)                                                       // Preserved
		assert.Equal(t, []string{"array"}, p2Schema.Type)                                                                             // Preserved
		assert.Equal(t, "csv", p2Schema.Format)                                                                                       // Preserved
		assert.Equal(t, []*yaml.Node{{Kind: yaml.ScalarNode, Value: "alpha"}, {Kind: yaml.ScalarNode, Value: "beta"}}, p2Schema.Enum) // Preserved
		require.NotNil(t, p2Schema.Default)                                                                                           // Preserved
		assert.Equal(t, "beta", p2Schema.Default.Value)                                                                               // Preserved
		require.NotNil(t, p2Schema.Items, "Schema Items should be present for p2")                                                    // Preserved
		assert.Equal(t, []string{"string"}, p2Schema.Items.A.Schema().Type)                                                           // Preserved
	})

	t.Run("appends new parameter if not existing", func(t *testing.T) {
		existingParams := []*v3.Parameter{
			{Name: "id", In: "path"},
		}
		newParam := &v3.Parameter{Name: "limit", In: "query", Schema: base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}})}

		updatedParams := MergeOrAppendParameter(existingParams, newParam)
		assert.Len(t, updatedParams, 2)
		assert.Equal(t, "limit", updatedParams[1].Name)
	})
}

func TestMergeParameters(t *testing.T) {
	t.Run("merges multiple parameters efficiently", func(t *testing.T) {
		existingParams := []*v3.Parameter{
			{Name: "p1", In: "query", Description: "old desc"},
			{Name: "p2", In: "query"},
		}
		newParams := []*v3.Parameter{
			{Name: "p1", In: "query", Description: "new desc"}, // Should not overwrite
			{Name: "p3", In: "query", Description: "added"},
			{Name: "p2", In: "query", Required: BoolPtr(true)},
		}

		updatedParams := MergeParameters(existingParams, newParams)

		assert.Len(t, updatedParams, 3)
		// Check order and content
		// p1
		assert.Equal(t, "p1", updatedParams[0].Name)
		assert.Equal(t, "old desc", updatedParams[0].Description)

		// p2
		assert.Equal(t, "p2", updatedParams[1].Name)
		assert.True(t, *updatedParams[1].Required)

		// p3
		assert.Equal(t, "p3", updatedParams[2].Name)
		assert.Equal(t, "added", updatedParams[2].Description)
	})

	t.Run("deduplicates new parameters when existing is empty", func(t *testing.T) {
		var existingParams []*v3.Parameter
		newParams := []*v3.Parameter{
			{Name: "p1", In: "query", Description: "first"},
			{Name: "p1", In: "query", Description: "second"}, // Should be merged into first
		}

		updatedParams := MergeParameters(existingParams, newParams)

		assert.Len(t, updatedParams, 1)
		assert.Equal(t, "p1", updatedParams[0].Name)
		// logic says: if p.Description == "" && newParam.Description != "" { ... }
		// so if first one has description, it keeps it. The only caveat to this is
		// for inferred * path parameters, since they have an auto-generated description.
		assert.Equal(t, "first", updatedParams[0].Description)
	})
}

func TestFilterInternalComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "no internal comments",
			input:    "regular comment",
			expected: "regular comment",
		},
		{
			name:     "with internal comment",
			input:    "regular comment (-- internal --)",
			expected: "regular comment",
		},
		{
			name:     "only internal comment",
			input:    "(-- internal --)",
			expected: "",
		},
		{
			name:     "multiline internal comment",
			input:    "start (-- \n internal \n --) end",
			expected: "start  end",
		},
		{
			name:     "multiple internal comments",
			input:    "one (-- two --) three (-- four --)",
			expected: "one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterInternalComments(tt.input); got != tt.expected {
				t.Errorf("filterInternalComments() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func BenchmarkFilterInternalComments(b *testing.B) {
	input := "Some public comment (-- internal comment --) more public comment"
	for b.Loop() {
		filterInternalComments(input)
	}
}

func TestAppendStringDedupe(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		toAppend string
		expected []string
	}{
		{
			name:     "empty list",
			input:    []string{},
			toAppend: "a",
			expected: []string{"a"},
		},
		{
			name:     "append new",
			input:    []string{"a", "b"},
			toAppend: "c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "append existing",
			input:    []string{"a", "b"},
			toAppend: "a",
			expected: []string{"a", "b"},
		},
		{
			name:     "append existing last",
			input:    []string{"a", "b"},
			toAppend: "b",
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := AppendStringDedupe(tt.input, tt.toAppend)
			if len(res) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(res))
			}
			for i := range res {
				if res[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, res)
				}
			}
		})
	}
}

func TestConvertNodeV3toV4(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		assert.Nil(t, ConvertNodeV3toV4(nil))
	})

	t.Run("scalar node", func(t *testing.T) {
		v3Node := &yamlv3.Node{
			Kind:        yamlv3.ScalarNode,
			Style:       yamlv3.DoubleQuotedStyle,
			Tag:         "!!str",
			Value:       "test-val",
			Anchor:      "anchor1",
			HeadComment: "head",
			LineComment: "line",
			FootComment: "foot",
			Line:        10,
			Column:      5,
		}
		v4Node := ConvertNodeV3toV4(v3Node)
		require.NotNil(t, v4Node)
		assert.Equal(t, yaml.ScalarNode, v4Node.Kind)
		assert.Equal(t, yaml.DoubleQuotedStyle, v4Node.Style)
		assert.Equal(t, "!!str", v4Node.Tag)
		assert.Equal(t, "test-val", v4Node.Value)
		assert.Equal(t, "anchor1", v4Node.Anchor)
		assert.Equal(t, "head", v4Node.HeadComment)
		assert.Equal(t, "line", v4Node.LineComment)
		assert.Equal(t, "foot", v4Node.FootComment)
		assert.Equal(t, 10, v4Node.Line)
		assert.Equal(t, 5, v4Node.Column)
	})

	t.Run("sequence node with alias and content", func(t *testing.T) {
		v3Node := &yamlv3.Node{
			Kind: yamlv3.SequenceNode,
			Alias: &yamlv3.Node{
				Kind:  yamlv3.ScalarNode,
				Value: "alias-val",
			},
			Content: []*yamlv3.Node{
				{Kind: yamlv3.ScalarNode, Value: "elem1"},
				{Kind: yamlv3.ScalarNode, Value: "elem2"},
			},
		}
		v4Node := ConvertNodeV3toV4(v3Node)
		require.NotNil(t, v4Node)
		assert.Equal(t, yaml.SequenceNode, v4Node.Kind)
		require.NotNil(t, v4Node.Alias)
		assert.Equal(t, "alias-val", v4Node.Alias.Value)
		require.Len(t, v4Node.Content, 2)
		assert.Equal(t, "elem1", v4Node.Content[0].Value)
		assert.Equal(t, "elem2", v4Node.Content[1].Value)
	})
}

func TestAppendComponents(t *testing.T) {
	spec := &v3.Document{
		Components: &v3.Components{
			Schemas:         orderedmap.New[string, *base.SchemaProxy](),
			Responses:       orderedmap.New[string, *v3.Response](),
			Parameters:      orderedmap.New[string, *v3.Parameter](),
			Examples:        orderedmap.New[string, *base.Example](),
			RequestBodies:   orderedmap.New[string, *v3.RequestBody](),
			Headers:         orderedmap.New[string, *v3.Header](),
			SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
			Links:           orderedmap.New[string, *v3.Link](),
			Callbacks:       orderedmap.New[string, *v3.Callback](),
		},
	}

	comp := &v3.Components{
		Schemas:         orderedmap.New[string, *base.SchemaProxy](),
		Responses:       orderedmap.New[string, *v3.Response](),
		Parameters:      orderedmap.New[string, *v3.Parameter](),
		Examples:        orderedmap.New[string, *base.Example](),
		RequestBodies:   orderedmap.New[string, *v3.RequestBody](),
		Headers:         orderedmap.New[string, *v3.Header](),
		SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
		Links:           orderedmap.New[string, *v3.Link](),
		Callbacks:       orderedmap.New[string, *v3.Callback](),
	}

	comp.Schemas.Set("MySchema", base.CreateSchemaProxy(&base.Schema{Title: "MySchema"}))
	comp.Responses.Set("200OK", &v3.Response{Description: "OK"})
	comp.Parameters.Set("limit", &v3.Parameter{Name: "limit"})
	comp.Examples.Set("Example1", &base.Example{Summary: "Example 1"})
	comp.RequestBodies.Set("MyBody", &v3.RequestBody{Description: "Body"})
	comp.Headers.Set("X-Trace", &v3.Header{Description: "Trace"})
	comp.SecuritySchemes.Set("OAuth2", &v3.SecurityScheme{Type: "oauth2"})
	comp.Links.Set("MyLink", &v3.Link{OperationId: "op1"})
	comp.Callbacks.Set("MyCallback", &v3.Callback{})

	AppendComponents(spec, comp)

	assert.Equal(t, 1, spec.Components.Schemas.Len())
	assert.Equal(t, 1, spec.Components.Responses.Len())
	assert.Equal(t, 1, spec.Components.Parameters.Len())
	assert.Equal(t, 1, spec.Components.Examples.Len())
	assert.Equal(t, 1, spec.Components.RequestBodies.Len())
	assert.Equal(t, 1, spec.Components.Headers.Len())
	assert.Equal(t, 1, spec.Components.SecuritySchemes.Len())
	assert.Equal(t, 1, spec.Components.Links.Len())
	assert.Equal(t, 1, spec.Components.Callbacks.Len())
}

func TestMergePathItems(t *testing.T) {
	existing := &v3.PathItem{
		Summary:     "Old summary",
		Description: "Old desc",
		Get: &v3.Operation{
			Summary: "Existing GET",
		},
		Extensions: orderedmap.New[string, *yaml.Node](),
	}
	newPI := &v3.PathItem{
		Summary:     "New summary",
		Description: "New desc",
		Get: &v3.Operation{
			Description: "New GET desc",
		},
		Post: &v3.Operation{
			Summary: "New POST",
		},
		Put: &v3.Operation{
			Summary: "New PUT",
		},
		Delete: &v3.Operation{
			Summary: "New DELETE",
		},
		Options: &v3.Operation{
			Summary: "New OPTIONS",
		},
		Head: &v3.Operation{
			Summary: "New HEAD",
		},
		Patch: &v3.Operation{
			Summary: "New PATCH",
		},
		Trace: &v3.Operation{
			Summary: "New TRACE",
		},
		Servers: []*v3.Server{
			{URL: "https://api.example.com"},
		},
		Extensions: orderedmap.New[string, *yaml.Node](),
	}
	newPI.Extensions.Set("x-custom", &yaml.Node{Value: "custom-val"})

	MergePathItems(existing, newPI)

	assert.Equal(t, "New summary", existing.Summary)
	assert.Equal(t, "New desc", existing.Description)
	assert.Equal(t, "Existing GET", existing.Get.Summary)
	assert.Equal(t, "New GET desc", existing.Get.Description)
	assert.NotNil(t, existing.Post)
	assert.NotNil(t, existing.Put)
	assert.NotNil(t, existing.Delete)
	assert.NotNil(t, existing.Options)
	assert.NotNil(t, existing.Head)
	assert.NotNil(t, existing.Patch)
	assert.NotNil(t, existing.Trace)
	assert.Len(t, existing.Servers, 1)
	assert.Equal(t, 1, existing.Extensions.Len())
}

func TestMergeResponses(t *testing.T) {
	t.Run("nil safety", func(t *testing.T) {
		MergeResponses(nil, nil)
		MergeResponses(&v3.Responses{}, nil)
		MergeResponses(nil, &v3.Responses{})
	})

	t.Run("merges codes and default response", func(t *testing.T) {
		existing := &v3.Responses{
			Codes: orderedmap.New[string, *v3.Response](),
		}
		existing.Codes.Set("200", &v3.Response{
			Description: "Old 200",
			Content:     orderedmap.New[string, *v3.MediaType](),
		})

		newResponses := &v3.Responses{
			Codes:   orderedmap.New[string, *v3.Response](),
			Default: &v3.Response{Description: "Default response"},
		}
		new200 := &v3.Response{
			Description: "New 200",
			Content:     orderedmap.New[string, *v3.MediaType](),
			Headers:     orderedmap.New[string, *v3.Header](),
			Links:       orderedmap.New[string, *v3.Link](),
			Extensions:  orderedmap.New[string, *yaml.Node](),
		}
		new200.Content.Set("application/json", &v3.MediaType{})
		new200.Headers.Set("X-RateLimit", &v3.Header{Description: "rate limit"})
		new200.Links.Set("UserLink", &v3.Link{OperationId: "getUser"})
		new200.Extensions.Set("x-code", &yaml.Node{Value: "123"})
		newResponses.Codes.Set("200", new200)
		newResponses.Codes.Set("404", &v3.Response{Description: "Not Found"})

		MergeResponses(existing, newResponses)

		assert.Equal(t, 2, existing.Codes.Len())
		resp200, ok := existing.Codes.Get("200")
		require.True(t, ok)
		assert.Equal(t, "New 200", resp200.Description)
		assert.Equal(t, 1, resp200.Content.Len())
		assert.Equal(t, 1, resp200.Headers.Len())
		assert.Equal(t, 1, resp200.Links.Len())
		assert.Equal(t, 1, resp200.Extensions.Len())
		assert.NotNil(t, existing.Default)
		assert.Equal(t, "Default response", existing.Default.Description)
	})
}

func TestSingular(t *testing.T) {
	tests := []struct {
		plural   string
		expected string
	}{
		{"calves", "calf"},
		{"knives", "knif"},
		{"categories", "category"},
		{"companies", "company"},
		{"users", "user"},
		{"orders", "order"},
		{"data", "data"},
	}

	for _, tt := range tests {
		t.Run(tt.plural, func(t *testing.T) {
			assert.Equal(t, tt.expected, Singular(tt.plural))
		})
	}
}

func TestResolveSchemaRef(t *testing.T) {
	assert.Equal(t, "#/components/schemas/foo.v1.User", ResolveSchemaRef(".foo.v1.User"))
	assert.Equal(t, "#/components/schemas/User", ResolveSchemaRef("#/components/schemas/User"))
	assert.Equal(t, "string", ResolveSchemaRef("string"))
}

func TestFormatTypeRef(t *testing.T) {
	assert.Equal(t, "foo.v1.User", FormatTypeRef(".foo.v1.User"))
	assert.Equal(t, "User", FormatTypeRef("User"))
}

func TestBoolPtr(t *testing.T) {
	b := BoolPtr(true)
	require.NotNil(t, b)
	assert.True(t, *b)
}

func TestMakePath(t *testing.T) {
	opts := options.Options{PathPrefix: "/api/v1"}
	assert.Equal(t, "/api/v1/users", MakePath(opts, "/users"))
	assert.Equal(t, "/api/v1/users", MakePath(opts, "users"))
}

