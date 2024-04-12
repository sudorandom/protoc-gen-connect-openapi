package gnostic

import (
	"log/slog"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/yaml.v2"
)

func SchemaWithSchemaAnnotations(schema *jsonschema.Schema, desc protoreflect.MessageDescriptor) *jsonschema.Schema {
	if !proto.HasExtension(desc.Options(), goa3.E_Schema.TypeDescriptor().Type()) {
		return schema
	}

	ext := proto.GetExtension(desc.Options(), goa3.E_Schema.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Schema)
	if !ok {
		return schema
	}
	return schemaWithAnnotations(schema, opts)
}

func SchemaWithPropertyAnnotations(schema *jsonschema.Schema, desc protoreflect.FieldDescriptor) *jsonschema.Schema {
	if !proto.HasExtension(desc.Options(), goa3.E_Property.TypeDescriptor().Type()) {
		return schema
	}

	ext := proto.GetExtension(desc.Options(), goa3.E_Property.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Schema)
	if !ok {
		return schema
	}
	return schemaWithAnnotations(schema, opts)
}

//gocyclo:ignore
func schemaWithAnnotations(schema *jsonschema.Schema, opts *goa3.Schema) *jsonschema.Schema {
	if opts.Description != "" {
		schema.Description = &opts.Description
	}
	if opts.Title != "" {
		schema.Title = &opts.Title
	}
	if opts.Format != "" {
		schema.Format = &opts.Format
	}
	if opts.Nullable {
		schema.WithExtraPropertiesItem("nullable", opts.Nullable)
	}
	if opts.ReadOnly {
		schema.ReadOnly = &opts.ReadOnly
	}
	if opts.WriteOnly {
		schema.WithExtraPropertiesItem("writeOnly", opts.WriteOnly)
	}
	if opts.Example != nil {
		// If the example is defined with the YAML option
		if opts.Example.Yaml != "" {
			var v interface{}
			if err := yaml.Unmarshal([]byte(opts.Example.GetYaml()), &v); err != nil {
				slog.Warn("unable to unmarshal example", slog.Any("error", err))
			} else {
				schema.Examples = append(schema.Examples, []interface{}{v})
			}
		}
		// If the example is defined with google.protobuf.Any
		if opts.Example.Value != nil {
			m, err := anypb.UnmarshalNew(opts.Example.GetValue(), proto.UnmarshalOptions{})
			if err != nil {
				slog.Warn("unable to unmarshal example", slog.Any("error", err))
			} else {
				schema.Examples = append(schema.Examples, []interface{}{m})
			}
		}
	}
	if opts.ExternalDocs != nil {
		schema.WithExtraPropertiesItem("externalDocs", toExternalDocs(opts.ExternalDocs))
	}
	if opts.Deprecated {
		schema.WithExtraPropertiesItem("deprecated", opts.Deprecated)
	}
	if opts.MultipleOf != 0 {
		schema.MultipleOf = &opts.MultipleOf
	}
	if opts.Maximum != 0 {
		if opts.ExclusiveMaximum {
			schema.ExclusiveMaximum = &opts.Maximum
		} else {
			schema.Maximum = &opts.Maximum
		}
	}
	if opts.Minimum != 0 {
		if opts.ExclusiveMinimum {
			schema.ExclusiveMinimum = &opts.Minimum
		} else {
			schema.Minimum = &opts.Minimum
		}
	}
	if opts.MaxLength > 0 {
		schema.MaxLength = &opts.MaxLength
	}
	if opts.MinLength > 0 {
		schema.MinLength = opts.MinLength
	}
	if opts.Pattern != "" {
		schema.Pattern = &opts.Pattern
	}
	if opts.MaxItems > 0 {
		schema.MaxItems = &opts.MaxItems
	}
	if opts.MinItems > 0 {
		schema.MinItems = opts.MinItems
	}
	if opts.UniqueItems {
		schema.UniqueItems = &opts.UniqueItems
	}
	if opts.MaxProperties > 0 {
		schema.MaxProperties = &opts.MaxProperties
	}
	if opts.MinProperties > 0 {
		schema.MinProperties = opts.MinProperties
	}
	if len(opts.Required) > 0 {
		schema.Required = opts.Required
	}
	if len(opts.Enum) > 0 {
		schema.Enum = []interface{}{opts.Enum}
	}
	if opts.Type != "" {
		t := jsonschema.SimpleType(opts.Type)
		schema.Type = &jsonschema.Type{SimpleTypes: &t}
	}

	if opts.AdditionalProperties != nil {
		switch v := opts.AdditionalProperties.GetOneof().(type) {
		case *goa3.AdditionalPropertiesItem_SchemaOrReference:
			vv := toSchemaOrBool(v.SchemaOrReference)
			schema.AdditionalProperties = &vv
		case *goa3.AdditionalPropertiesItem_Boolean:
			schema.AdditionalItems = &jsonschema.SchemaOrBool{TypeBoolean: &v.Boolean}
		}
	}

	if len(opts.AllOf) > 0 {
		schema.AllOf = toSchemaOrBools(opts.AllOf)
	}
	if len(opts.OneOf) > 0 {
		schema.OneOf = toSchemaOrBools(opts.OneOf)
	}
	if len(opts.AnyOf) > 0 {
		schema.AnyOf = toSchemaOrBools(opts.AnyOf)
	}
	if opts.Not != nil {
		v := toSchemaOrBool(&goa3.SchemaOrReference{
			Oneof: &goa3.SchemaOrReference_Schema{
				Schema: opts.Not,
			},
		})
		schema.Not = &v
	}
	if opts.Items != nil {
		schema.Items = &jsonschema.Items{
			SchemaOrBool: &jsonschema.SchemaOrBool{},
			SchemaArray:  toSchemaOrBools(opts.Items.GetSchemaOrReference()),
		}
	}
	if opts.Properties != nil {
		schema.Properties = toSchemaOrBoolMap(opts.Properties.GetAdditionalProperties())
	}
	if opts.Default != nil {
		schema.Default = toDefault(opts.Default)
	}
	if opts.AdditionalProperties != nil {
		schema.AdditionalProperties = toAdditionalPropertiesItem(opts.AdditionalProperties)
	}

	return schema
}
