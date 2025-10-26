package gnostic

import (
	"fmt"
	"log/slog"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SchemaWithSchemaAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	if !proto.HasExtension(desc.Options(), goa3.E_Schema.TypeDescriptor().Type()) {
		return schema
	}

	ext := proto.GetExtension(desc.Options(), goa3.E_Schema.TypeDescriptor().Type())
	gnosticSchema, ok := ext.(*goa3.Schema)
	if !ok {
		return schema
	}
	return schemaWithAnnotations(opts, schema, gnosticSchema)
}

func SchemaWithPropertyAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	if !proto.HasExtension(desc.Options(), goa3.E_Property.TypeDescriptor().Type()) {
		return schema
	}

	ext := proto.GetExtension(desc.Options(), goa3.E_Property.TypeDescriptor().Type())
	gnosticSchema, ok := ext.(*goa3.Schema)
	if !ok {
		return schema
	}
	return schemaWithAnnotations(opts, schema, gnosticSchema)
}

//gocyclo:ignore
func schemaWithAnnotations(opts options.Options, schema *base.Schema, gnosticSchema *goa3.Schema) *base.Schema {
	if gnosticSchema.Description != "" {
		schema.Description = gnosticSchema.Description
	}
	if gnosticSchema.Title != "" {
		schema.Title = gnosticSchema.Title
	}
	if gnosticSchema.Format != "" {
		schema.Format = gnosticSchema.Format
	}
	if gnosticSchema.Nullable {
		schema.Nullable = &gnosticSchema.Nullable
	}
	if gnosticSchema.ReadOnly {
		schema.ReadOnly = &gnosticSchema.ReadOnly
	}
	if gnosticSchema.WriteOnly {
		schema.WriteOnly = &gnosticSchema.WriteOnly
	}
	if gnosticSchema.Example != nil {
		// If the example is defined with the YAML option
		if gnosticSchema.Example.Yaml != "" {
			var v string
			if err := yaml.Unmarshal([]byte(gnosticSchema.Example.GetYaml()), &v); err != nil {
				var node any
				if err := yaml.Unmarshal([]byte(gnosticSchema.Example.GetYaml()), &node); err != nil {
					opts.Logger.Warn("unable to unmarshal example", slog.Any("error", err))
				} else {
					schema.Examples = append(schema.Examples, expandExampleEntry(node))
				}
			} else {
				schema.Examples = append(schema.Examples, utils.CreateStringNode(v))
			}
		}
		// If the example is defined with google.protobuf.Any
		if gnosticSchema.Example.Value != nil {
			opts.Logger.Warn("unable to unmarshal pb.any example")
		}
	}
	if gnosticSchema.ExternalDocs != nil {
		schema.ExternalDocs = toExternalDocs(gnosticSchema.ExternalDocs)
	}
	if gnosticSchema.Deprecated {
		schema.Deprecated = &gnosticSchema.Deprecated
	}
	if gnosticSchema.MultipleOf != 0 {
		schema.MultipleOf = &gnosticSchema.MultipleOf
	}
	if gnosticSchema.Maximum != 0 {
		if gnosticSchema.ExclusiveMaximum {
			schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: gnosticSchema.Maximum}
		} else {
			schema.Maximum = &gnosticSchema.Maximum
		}
	}
	if gnosticSchema.Minimum != 0 {
		if gnosticSchema.ExclusiveMinimum {
			schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: gnosticSchema.Minimum}
		} else {
			schema.Minimum = &gnosticSchema.Minimum
		}
	}
	if gnosticSchema.MaxLength > 0 {
		schema.MaxLength = &gnosticSchema.MaxLength
	}
	if gnosticSchema.MinLength > 0 {
		v := gnosticSchema.MinLength
		schema.MinLength = &v
	}
	if gnosticSchema.Pattern != "" {
		schema.Pattern = gnosticSchema.Pattern
	}
	if gnosticSchema.MaxItems > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MaxItems = &gnosticSchema.MaxItems
		}
	}
	if gnosticSchema.MinItems > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MinItems = &gnosticSchema.MinItems
		}
	}
	if gnosticSchema.UniqueItems {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().UniqueItems = &gnosticSchema.UniqueItems
		}
	}
	if gnosticSchema.MaxProperties > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MaxProperties = &gnosticSchema.MaxProperties
		}
	}
	if gnosticSchema.MinProperties > 0 {
		if schema.ParentProxy != nil {
			schema.ParentProxy.Schema().MinProperties = &gnosticSchema.MinProperties
		}
	}
	if len(gnosticSchema.Required) > 0 {
		schema.Required = gnosticSchema.Required
	}
	if len(gnosticSchema.Enum) > 0 {
		enums := make([]*yaml.Node, len(gnosticSchema.Enum))
		for i, enum := range gnosticSchema.Enum {
			enums[i] = util.ConvertNodeV3toV4(enum.ToRawInfo())
		}
		schema.Enum = enums
	}
	if gnosticSchema.Type != "" {
		schema.Type = []string{gnosticSchema.Type}
	}

	if gnosticSchema.AdditionalProperties != nil {
		switch v := gnosticSchema.AdditionalProperties.GetOneof().(type) {
		case *goa3.AdditionalPropertiesItem_SchemaOrReference:
			if vv := toSchemaOrReference(opts, v.SchemaOrReference); vv != nil {
				schema.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{A: vv}
			}
		case *goa3.AdditionalPropertiesItem_Boolean:
			schema.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: v.Boolean}
		}
	}

	if len(gnosticSchema.AllOf) > 0 {
		schema.AllOf = toSchemaOrReferences(opts, gnosticSchema.AllOf)
	}
	if len(gnosticSchema.OneOf) > 0 {
		schema.OneOf = toSchemaOrReferences(opts, gnosticSchema.OneOf)
	}
	if len(gnosticSchema.AnyOf) > 0 {
		schema.AnyOf = toSchemaOrReferences(opts, gnosticSchema.AnyOf)
	}
	if gnosticSchema.Not != nil {
		schema.Not = base.CreateSchemaProxy(toSchema(opts, gnosticSchema.Not))
	}
	if gnosticSchema.Items != nil {
		items := toSchemaOrReferences(opts, gnosticSchema.Items.SchemaOrReference)
		var itemsSchema *base.SchemaProxy
		if len(items) == 1 {
			itemsSchema = items[0]
		} else {
			itemsSchema = base.CreateSchemaProxy(&base.Schema{OneOf: items})
		}
		schema.Items = &base.DynamicValue[*base.SchemaProxy, bool]{A: itemsSchema}
	}
	if gnosticSchema.Properties != nil {
		schema.Properties = toSchemaOrReferenceMap(opts, gnosticSchema.Properties.GetAdditionalProperties())
	}
	if gnosticSchema.Default != nil {
		schema.Default = toDefault(gnosticSchema.Default)
	}
	if gnosticSchema.AdditionalProperties != nil {
		schema.AdditionalProperties = toAdditionalPropertiesItem(opts, gnosticSchema.AdditionalProperties)
	}
	if gnosticSchema.Xml != nil {
		extensions := *orderedmap.New[string, *yaml.Node]()
		for _, namedAny := range gnosticSchema.Xml.GetSpecificationExtension() {
			extensions.Set(namedAny.Name, util.ConvertNodeV3toV4(namedAny.ToRawInfo()))
		}
		schema.XML = &base.XML{
			Name:       gnosticSchema.Xml.Name,
			Namespace:  gnosticSchema.Xml.Namespace,
			Prefix:     gnosticSchema.Xml.Prefix,
			Attribute:  gnosticSchema.Xml.Attribute,
			Wrapped:    gnosticSchema.Xml.Wrapped,
			Extensions: &extensions,
		}
	}
	if gnosticSchema.Discriminator != nil {
		mapping := orderedmap.New[string, string]()
		for _, prop := range gnosticSchema.Discriminator.GetMapping().GetAdditionalProperties() {
			mapping.Set(prop.Name, prop.Value)
		}
		schema.Discriminator = &base.Discriminator{
			PropertyName: gnosticSchema.Discriminator.GetPropertyName(),
			Mapping:      mapping,
		}
	}
	if gnosticSchema.SpecificationExtension != nil {
		schema.Extensions = toExtensions(gnosticSchema.SpecificationExtension)
	}

	return schema
}

func expandExampleEntry(entry any) *yaml.Node {
	switch vv := entry.(type) {
	//Handle mapping nodes
	case map[any]any:
		node := utils.CreateEmptyMapNode()
		for k, v := range vv {
			keyNode := utils.CreateStringNode(k.(string))
			node.Content = append(node.Content, keyNode, expandExampleEntry(v))
		}
		return node
	// Handle sequence nodes
	case []any:
		node := utils.CreateEmptySequenceNode()
		for _, v := range vv {
			node.Content = append(node.Content, expandExampleEntry(v))
		}
		return node
	case float64:
		return utils.CreateFloatNode(fmt.Sprintf("%v", vv))
	case int:
		return utils.CreateIntNode(fmt.Sprintf("%v", vv))
	case string:
		return utils.CreateStringNode(vv)
	default:
		return utils.CreateStringNode(fmt.Sprintf("%v", vv))
	}
}
