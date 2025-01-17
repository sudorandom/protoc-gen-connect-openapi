package googleapi

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SchemaWithPropertyAnnotations(schema *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	dopts := desc.Options()
	if !proto.HasExtension(dopts, annotations.E_FieldBehavior) {
		return schema
	}
	fieldBehavior, ok := proto.GetExtension(dopts, annotations.E_FieldBehavior).(*annotations.FieldBehavior)
	if !ok {
		return schema
	}
	fb := fieldBehavior.Enum()
	if fb == nil {
		return schema
	}
	switch *fb {
	case annotations.FieldBehavior_FIELD_BEHAVIOR_UNSPECIFIED:
	case annotations.FieldBehavior_OPTIONAL:
		schema.Description = "(OPTIONAL) " + schema.Description
	case annotations.FieldBehavior_REQUIRED:
		schema.ParentProxy.Schema().Required = util.AppendStringDedupe(schema.ParentProxy.Schema().Required, string(desc.Name()))
	case annotations.FieldBehavior_OUTPUT_ONLY:
		schema.ReadOnly = util.BoolPtr(true)
	case annotations.FieldBehavior_INPUT_ONLY:
		schema.WriteOnly = util.BoolPtr(true)
	case annotations.FieldBehavior_IMMUTABLE:
		schema.Description = "(IMMUTABLE) " + schema.Description
	case annotations.FieldBehavior_UNORDERED_LIST:
		schema.Description = "(UNORDERED_LIST) " + schema.Description
	case annotations.FieldBehavior_NON_EMPTY_DEFAULT:
		schema.Description = "(NON_EMPTY_DEFAULT) " + schema.Description
	case annotations.FieldBehavior_IDENTIFIER:
		schema.Description = "(IDENTIFIER) " + schema.Description
	default:
	}
	return schema
}
