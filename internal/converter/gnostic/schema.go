package gnostic

import (
	"log/slog"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		schema.ExtraProperties["nullable"] = opts.Nullable
	}
	if opts.ReadOnly {
		schema.ReadOnly = &opts.ReadOnly
	}
	if opts.WriteOnly {
		schema.ExtraProperties["writeOnly"] = opts.WriteOnly
	}
	if opts.Example != nil {
		schema.Examples = []interface{}{opts.Example}
	}
	if opts.ExternalDocs != nil {
		schema.ExtraProperties["externalDocs"] = toExternalDocs(opts.ExternalDocs)
	}
	if opts.Deprecated {
		schema.ExtraProperties["deprecated"] = opts.Deprecated
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

	if opts.AdditionalProperties.GetBoolean() {
		b := opts.AdditionalProperties.GetBoolean()
		schema.AdditionalItems = &jsonschema.SchemaOrBool{
			TypeBoolean: &b,
		}
	} else if s := opts.AdditionalProperties.GetSchemaOrReference(); s != nil {
		slog.Warn("additional_properties with a schema is not yet supported", slog.Any("additional_properties", opts.AdditionalProperties))
	}

	if len(opts.AllOf) > 0 {
		slog.Warn("all_of is not supported", slog.Any("all_of", opts.AllOf))
	}
	if len(opts.OneOf) > 0 {
		slog.Warn("one_of is not supported", slog.Any("one_of", opts.OneOf))
	}
	if len(opts.AnyOf) > 0 {
		slog.Warn("any_of is not supported", slog.Any("any_of", opts.AnyOf))
	}
	if opts.Not != nil {
		slog.Warn("not is not supported", slog.Any("items", opts.Not))
	}
	if opts.Items != nil {
		slog.Warn("items is not supported", slog.Any("items", opts.Items))
	}
	if opts.Properties != nil {
		slog.Warn("properties is not supported", slog.Any("properties", opts.Properties))
	}
	if opts.Default != nil {
		slog.Warn("default is not supported", slog.Any("default", opts.Default))
	}

	return schema
}
