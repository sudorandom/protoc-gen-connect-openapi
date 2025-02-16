package gnostic

import (
	"log/slog"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

func SchemaWithSchemaAnnotations(schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
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

func SchemaWithPropertyAnnotations(schema *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
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
func schemaWithAnnotations(schema *base.Schema, opts *goa3.Schema) *base.Schema {
	if opts.Description != "" {
		schema.Description = opts.Description
	}
	if opts.Title != "" {
		schema.Title = opts.Title
	}
	if opts.Format != "" {
		schema.Format = opts.Format
	}
	if opts.Nullable {
		schema.Nullable = &opts.Nullable
	}
	if opts.ReadOnly {
		schema.ReadOnly = &opts.ReadOnly
	}
	if opts.WriteOnly {
		schema.WriteOnly = &opts.WriteOnly
	}
	if opts.Example != nil {
		// If the example is defined with the YAML option
		if opts.Example.Yaml != "" {
			var v string
			if err := yaml.Unmarshal([]byte(opts.Example.GetYaml()), &v); err != nil {
				var node yaml.Node
				if err := yaml.Unmarshal([]byte(opts.Example.GetYaml()), &node); err != nil {
					slog.Warn("unable to unmarshal example", slog.Any("error", err))
				} else {
					schema.Examples = append(schema.Examples, &node)
				}
			} else {
				schema.Examples = append(schema.Examples, utils.CreateStringNode(v))
			}
		}
		// If the example is defined with google.protobuf.Any
		if opts.Example.Value != nil {
			slog.Warn("unable to unmarshal pb.any example")
		}
	}
	if opts.ExternalDocs != nil {
		schema.ExternalDocs = toExternalDocs(opts.ExternalDocs)
	}
	if opts.Deprecated {
		schema.Deprecated = &opts.Deprecated
	}
	if opts.MultipleOf != 0 {
		schema.MultipleOf = &opts.MultipleOf
	}
	if opts.Maximum != 0 {
		if opts.ExclusiveMaximum {
			schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: opts.Maximum}
		} else {
			schema.Maximum = &opts.Maximum
		}
	}
	if opts.Minimum != 0 {
		if opts.ExclusiveMinimum {
			schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: opts.Minimum}
		} else {
			schema.Minimum = &opts.Minimum
		}
	}
	if opts.MaxLength > 0 {
		schema.MaxLength = &opts.MaxLength
	}
	if opts.MinLength > 0 {
		v := opts.MinLength
		schema.MinLength = &v
	}
	if opts.Pattern != "" {
		schema.Pattern = opts.Pattern
	}
	if opts.MaxItems > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MaxItems = &opts.MaxItems
		}
	}
	if opts.MinItems > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MinItems = &opts.MinItems
		}
	}
	if opts.UniqueItems {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().UniqueItems = &opts.UniqueItems
		}
	}
	if opts.MaxProperties > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MaxProperties = &opts.MaxProperties
		}
	}
	if opts.MinProperties > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MinProperties = &opts.MinProperties
		}
	}
	if len(opts.Required) > 0 {
		schema.Required = opts.Required
	}
	if len(opts.Enum) > 0 {
		enums := make([]*yaml.Node, len(opts.Enum))
		for i, enum := range opts.Enum {
			enums[i] = enum.ToRawInfo()
		}
		schema.Enum = enums
	}
	if opts.Type != "" {
		schema.Type = []string{opts.Type}
	}

	if opts.AdditionalProperties != nil {
		switch v := opts.AdditionalProperties.GetOneof().(type) {
		case *goa3.AdditionalPropertiesItem_SchemaOrReference:
			if vv := toSchemaOrReference(v.SchemaOrReference); vv != nil {
				schema.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{A: vv}
			}
		case *goa3.AdditionalPropertiesItem_Boolean:
			schema.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: v.Boolean}
		}
	}

	if len(opts.AllOf) > 0 {
		schema.AllOf = toSchemaOrReferences(opts.AllOf)
	}
	if len(opts.OneOf) > 0 {
		schema.OneOf = toSchemaOrReferences(opts.OneOf)
	}
	if len(opts.AnyOf) > 0 {
		schema.AnyOf = toSchemaOrReferences(opts.AnyOf)
	}
	if opts.Not != nil {
		schema.Not = base.CreateSchemaProxy(toSchema(opts.Not))
	}
	if opts.Items != nil {
		items := toSchemaOrReferences(opts.Items.SchemaOrReference)
		var itemsSchema *base.SchemaProxy
		if len(items) == 1 {
			itemsSchema = items[0]
		} else {
			itemsSchema = base.CreateSchemaProxy(&base.Schema{OneOf: items})
		}
		schema.Items = &base.DynamicValue[*base.SchemaProxy, bool]{A: itemsSchema}
	}
	if opts.Properties != nil {
		schema.Properties = toSchemaOrReferenceMap(opts.Properties.GetAdditionalProperties())
	}
	if opts.Default != nil {
		schema.Default = toDefault(opts.Default)
	}
	if opts.AdditionalProperties != nil {
		schema.AdditionalProperties = toAdditionalPropertiesItem(opts.AdditionalProperties)
	}
	if opts.Xml != nil {
		extensions := *orderedmap.New[string, *yaml.Node]()
		for _, namedAny := range opts.Xml.GetSpecificationExtension() {
			extensions.Set(namedAny.Name, namedAny.ToRawInfo())
		}
		schema.XML = &base.XML{
			Name:       opts.Xml.Name,
			Namespace:  opts.Xml.Namespace,
			Prefix:     opts.Xml.Prefix,
			Attribute:  opts.Xml.Attribute,
			Wrapped:    opts.Xml.Wrapped,
			Extensions: &extensions,
		}
	}
	if opts.Discriminator != nil {
		mapping := orderedmap.New[string, string]()
		for _, prop := range opts.Discriminator.GetMapping().GetAdditionalProperties() {
			mapping.Set(prop.Name, prop.Value)
		}
		schema.Discriminator = &base.Discriminator{
			PropertyName: opts.Discriminator.GetPropertyName(),
			Mapping:      mapping,
		}
	}
	if opts.SpecificationExtension != nil {
		schema.Extensions = toExtensions(opts.SpecificationExtension)
	}

	return schema
}
