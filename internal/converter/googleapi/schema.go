package googleapi

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// IsFieldRequired determines if a parameter should be required based on field behavior annotations.
// Returns:
//   - *bool(true) if field has REQUIRED behavior
//   - *bool(false) if field has OPTIONAL behavior (and not REQUIRED)
//   - nil if field has neither, allowing other factors to determine required status
func IsFieldRequired(desc protoreflect.FieldDescriptor) *bool {
	dopts := desc.Options()
	if !proto.HasExtension(dopts, annotations.E_FieldBehavior) {
		return nil
	}
	fieldBehaviors, ok := proto.GetExtension(dopts, annotations.E_FieldBehavior).([]annotations.FieldBehavior)
	if !ok {
		return nil
	}

	hasRequired := false
	hasOptional := false

	for _, fieldBehavior := range fieldBehaviors {
		fb := fieldBehavior.Enum()
		if fb == nil {
			continue
		}
		switch *fb {
		case annotations.FieldBehavior_REQUIRED:
			hasRequired = true
		case annotations.FieldBehavior_OPTIONAL:
			hasOptional = true
		}
	}

	// REQUIRED takes precedence over OPTIONAL
	if hasRequired {
		return util.BoolPtr(true)
	}
	if hasOptional {
		return util.BoolPtr(false)
	}
	return nil
}

func SchemaWithPropertyAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	dopts := desc.Options()
	if !proto.HasExtension(dopts, annotations.E_FieldBehavior) {
		return schema
	}
	fieldBehaviors, ok := proto.GetExtension(dopts, annotations.E_FieldBehavior).([]annotations.FieldBehavior)
	if !ok {
		return schema
	}
	for _, fieldBehavior := range fieldBehaviors {
		fb := fieldBehavior.Enum()
		if fb == nil {
			return schema
		}
		switch *fb {
		case annotations.FieldBehavior_FIELD_BEHAVIOR_UNSPECIFIED:
		case annotations.FieldBehavior_OPTIONAL:
			schema.Description = "(OPTIONAL) " + schema.Description
		case annotations.FieldBehavior_REQUIRED:
			if schema.ParentProxy != nil {
				schema.ParentProxy.Schema().Required = util.AppendStringDedupe(schema.ParentProxy.Schema().Required, util.MakeFieldName(opts, desc))
			}
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
	}
	return schema
}
