package gnostic

import (
	"testing"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestPathItemWithMethodAnnotations_ParameterMerging(t *testing.T) {
	t.Run("merges parameters from gnostic annotations", func(t *testing.T) {
		// Create existing path item with operation
		item := &v3.PathItem{
			Post: &v3.Operation{
				Parameters: []*v3.Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: "Original ID",
					},
					{
						Name: "filter",
						In:   "query",
					},
				},
			},
		}

		// Create method descriptor with gnostic operation annotation
		gnosticOp := &goa3.Operation{
			Parameters: []*goa3.ParameterOrReference{
				{
					Oneof: &goa3.ParameterOrReference_Parameter{
						Parameter: &goa3.Parameter{
							Name:        "id",
							In:          "path",
							Description: "ID from annotation",
							Required:    true,
						},
					},
				},
				{
					Oneof: &goa3.ParameterOrReference_Parameter{
						Parameter: &goa3.Parameter{
							Name: "limit",
							In:   "query",
						},
					},
				},
			},
		}

		methodOpts := &descriptorpb.MethodOptions{}
		proto.SetExtension(methodOpts, goa3.E_Operation, gnosticOp)

		md := &mockMethodDescriptor{
			options: methodOpts,
		}

		opts := options.Options{}
		result := PathItemWithMethodAnnotations(opts, item, md)

		// Verify parameters were merged
		require.NotNil(t, result.Post)
		assert.Len(t, result.Post.Parameters, 3, "should have 3 parameters after merge")

		// Check id parameter - should preserve original description
		idParam := result.Post.Parameters[0]
		assert.Equal(t, "id", idParam.Name)
		assert.Equal(t, "path", idParam.In)
		assert.Equal(t, "Original ID", idParam.Description, "should preserve original description")
		assert.True(t, *idParam.Required, "should merge required flag")

		// Check filter parameter - should be unchanged
		filterParam := result.Post.Parameters[1]
		assert.Equal(t, "filter", filterParam.Name)
		assert.Equal(t, "query", filterParam.In)

		// Check limit parameter - should be added
		limitParam := result.Post.Parameters[2]
		assert.Equal(t, "limit", limitParam.Name)
		assert.Equal(t, "query", limitParam.In)
	})

	t.Run("appends new parameters from annotations", func(t *testing.T) {
		item := &v3.PathItem{
			Get: &v3.Operation{
				Parameters: []*v3.Parameter{},
			},
		}

		gnosticOp := &goa3.Operation{
			Parameters: []*goa3.ParameterOrReference{
				{
					Oneof: &goa3.ParameterOrReference_Parameter{
						Parameter: &goa3.Parameter{
							Name:        "api_key",
							In:          "header",
							Description: "API Key",
							Required:    true,
						},
					},
				},
			},
		}

		methodOpts := &descriptorpb.MethodOptions{}
		proto.SetExtension(methodOpts, goa3.E_Operation, gnosticOp)

		md := &mockMethodDescriptor{
			options: methodOpts,
		}

		opts := options.Options{}
		result := PathItemWithMethodAnnotations(opts, item, md)

		require.NotNil(t, result.Get)
		assert.Len(t, result.Get.Parameters, 1)
		assert.Equal(t, "api_key", result.Get.Parameters[0].Name)
	})

	t.Run("returns unchanged item when no gnostic operation extension", func(t *testing.T) {
		item := &v3.PathItem{
			Post: &v3.Operation{
				Summary: "Original operation",
				Parameters: []*v3.Parameter{
					{Name: "id", In: "path"},
				},
			},
		}

		md := &mockMethodDescriptor{
			options: &descriptorpb.MethodOptions{},
		}

		opts := options.Options{}
		result := PathItemWithMethodAnnotations(opts, item, md)

		assert.Equal(t, item, result)
		assert.Len(t, result.Post.Parameters, 1)
		assert.Equal(t, "id", result.Post.Parameters[0].Name)
	})

	t.Run("merges parameters across multiple operations in path item", func(t *testing.T) {
		item := &v3.PathItem{
			Get: &v3.Operation{
				Parameters: []*v3.Parameter{
					{Name: "id", In: "path"},
				},
			},
			Post: &v3.Operation{
				Parameters: []*v3.Parameter{
					{Name: "id", In: "path"},
				},
			},
		}

		gnosticOp := &goa3.Operation{
			Parameters: []*goa3.ParameterOrReference{
				{
					Oneof: &goa3.ParameterOrReference_Parameter{
						Parameter: &goa3.Parameter{
							Name: "filter",
							In:   "query",
						},
					},
				},
			},
		}

		methodOpts := &descriptorpb.MethodOptions{}
		proto.SetExtension(methodOpts, goa3.E_Operation, gnosticOp)

		md := &mockMethodDescriptor{
			options: methodOpts,
		}

		opts := options.Options{}
		result := PathItemWithMethodAnnotations(opts, item, md)

		// Both GET and POST should have the filter parameter added
		require.NotNil(t, result.Get)
		assert.Len(t, result.Get.Parameters, 2)
		assert.Equal(t, "filter", result.Get.Parameters[1].Name)

		require.NotNil(t, result.Post)
		assert.Len(t, result.Post.Parameters, 2)
		assert.Equal(t, "filter", result.Post.Parameters[1].Name)
	})

	t.Run("preserves parameter schema details when merging", func(t *testing.T) {
		item := &v3.PathItem{
			Get: &v3.Operation{
				Parameters: []*v3.Parameter{
					{
						Name:        "page",
						In:          "query",
						Description: "Original page description",
						Schema: base.CreateSchemaProxy(&base.Schema{
							Type:    []string{"integer"},
							Default: &yaml.Node{Value: "1"},
						}),
					},
				},
			},
		}

		gnosticOp := &goa3.Operation{
			Parameters: []*goa3.ParameterOrReference{
				{
					Oneof: &goa3.ParameterOrReference_Parameter{
						Parameter: &goa3.Parameter{
							Name: "page",
							In:   "query",
							Schema: &goa3.SchemaOrReference{
								Oneof: &goa3.SchemaOrReference_Schema{
									Schema: &goa3.Schema{
										Type:   "integer",
										Format: "int32",
									},
								},
							},
						},
					},
				},
			},
		}

		methodOpts := &descriptorpb.MethodOptions{}
		proto.SetExtension(methodOpts, goa3.E_Operation, gnosticOp)

		md := &mockMethodDescriptor{
			options: methodOpts,
		}

		opts := options.Options{}
		result := PathItemWithMethodAnnotations(opts, item, md)

		require.NotNil(t, result.Get)
		assert.Len(t, result.Get.Parameters, 1)

		pageParam := result.Get.Parameters[0]
		assert.Equal(t, "page", pageParam.Name)
		assert.Equal(t, "Original page description", pageParam.Description, "should preserve original description")
		// The schema should have been merged with format from gnostic annotation
		assert.Equal(t, "int32", pageParam.Schema.Schema().Format)
	})
}

// Mock implementation of protoreflect.MethodDescriptor for testing
type mockMethodDescriptor struct {
	protoreflect.MethodDescriptor
	options *descriptorpb.MethodOptions
}

func (m *mockMethodDescriptor) Options() protoreflect.ProtoMessage {
	return m.options
}
