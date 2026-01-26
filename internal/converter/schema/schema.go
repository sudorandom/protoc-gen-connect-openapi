package schema

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/visibility"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageToSchema(opts options.Options, tt protoreflect.MessageDescriptor) (string, *base.Schema) {
	opts.Logger.Debug("messageToSchema", slog.Any("descriptor", tt.FullName()))
	defer opts.Logger.Debug("/messageToSchema", slog.Any("descriptor", tt.FullName()))
	if util.IsWellKnown(tt) {
		wk := util.WellKnownToSchema(tt)
		if wk == nil {
			return "", nil
		}
		return wk.ID, wk.Schema
	}
	title := string(tt.Name())
	if opts.FullyQualifiedMessageNames {
		title = string(tt.FullName())
	}
	s := &base.Schema{
		Title:                title,
		Description:          util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
		Type:                 []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
	}

	oneOneGroups := map[protoreflect.FullName][]protoreflect.FieldDescriptor{}
	regularProps := orderedmap.New[string, *base.SchemaProxy]()

	fields := tt.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(field), opts.AllowedVisibilities) {
			opts.Logger.Debug("Filtering field due to visibility", slog.String("field", string(field.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
			continue
		}
		if oneOf := field.ContainingOneof(); oneOf != nil && !oneOf.IsSynthetic() {
			oneOneGroups[oneOf.FullName()] = append(oneOneGroups[oneOf.FullName()], field)
			continue
		}
		prop := FieldToSchema(opts, base.CreateSchemaProxy(s), field)
		if field.HasOptionalKeyword() {
			schema := prop.Schema()
			if schema == nil {
				continue
			}

			switch field.Kind() {
			case protoreflect.MessageKind, protoreflect.EnumKind: // now handled in FieldToSchema
			default:
				appendType(schema, "null")
			}
		}
		regularProps.Set(util.MakeFieldName(opts, field), prop)
	}

	s.Properties = regularProps
	if len(oneOneGroups) > 0 {
		// make all of groups
		groupKeys := []protoreflect.FullName{}
		for key := range oneOneGroups {
			groupKeys = append(groupKeys, key)
		}
		slices.Sort(groupKeys)
		allOfs := []*base.SchemaProxy{}
		for _, key := range groupKeys {
			items := oneOneGroups[key]
			slices.SortFunc(items, func(a, b protoreflect.FieldDescriptor) int {
				return strings.Compare(string(a.Name()), string(b.Name()))
			})
			allOfs = append(allOfs, makeOneOfGroup(opts, items))
		}
		if len(allOfs) == 1 {
			s.OneOf = allOfs[0].Schema().OneOf
		} else {
			s.AllOf = append(s.AllOf, allOfs...)
		}
	}

	// if there are oneOfs and properties, we should merge them under a allOf.
	// having properties and oneOfs at the same level creates conflicts.
	if len(s.OneOf) > 0 && s.Properties.Len() > 0 {
		s.AllOf = append(s.AllOf,
			base.CreateSchemaProxy(&base.Schema{Properties: s.Properties}),
			base.CreateSchemaProxy(&base.Schema{OneOf: s.OneOf}),
		)
		s.Properties = nil
		s.OneOf = nil
	}

	// Apply Updates from Options
	s = opts.MessageAnnotator.AnnotateMessage(opts, s, tt)
	return string(tt.FullName()), s
}

func FieldToSchema(opts options.Options, parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.SchemaProxy {
	opts.Logger.Debug("FieldToSchema", slog.Any("descriptor", tt.FullName()))
	defer opts.Logger.Debug("/FieldToSchema", slog.Any("descriptor", tt.FullName()))

	if tt.IsMap() {
		// Handle maps
		root := ScalarFieldToSchema(opts, parent, tt, false)
		root.Title = string(tt.Name())
		root.Type = []string{"object"}
		root.Description = util.TypeFieldDescription(opts, tt)
		root.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{A: FieldToSchema(opts, parent, tt.MapValue())}
		root = opts.FieldAnnotator.AnnotateField(opts, root, tt, false)
		return base.CreateSchemaProxy(root)
	} else if tt.IsList() {
		var itemSchema *base.SchemaProxy
		switch tt.Kind() {
		case protoreflect.MessageKind:
			itemSchema = ReferenceFieldToSchema(opts, parent, tt)
		case protoreflect.EnumKind:
			itemSchema = ReferenceFieldToSchema(opts, parent, tt)
		default:
			itemSchema = base.CreateSchemaProxy(ScalarFieldToSchema(opts, parent, tt, true))
		}
		s := &base.Schema{
			Title:       string(tt.Name()),
			ParentProxy: parent,
			Description: util.TypeFieldDescription(opts, tt),
			Type:        []string{"array"},
			Items:       &base.DynamicValue[*base.SchemaProxy, bool]{A: itemSchema},
			Deprecated:  util.IsFieldDeprecated(tt),
		}
		s = opts.FieldAnnotator.AnnotateField(opts, s, tt, false)
		return base.CreateSchemaProxy(s)
	} else {
		switch tt.Kind() {
		case protoreflect.MessageKind, protoreflect.EnumKind:
			msg := ScalarFieldToSchema(opts, parent, tt, false)
			ref := ReferenceFieldToSchema(opts, parent, tt)
			if tt.HasOptionalKeyword() {
				msg.OneOf = []*base.SchemaProxy{
					ref,
					base.CreateSchemaProxy(&base.Schema{Type: []string{"null"}}),
				}
			} else {
				extensions := orderedmap.New[string, *yaml.Node]()
				extensions.Set("$ref", utils.CreateStringNode(ref.GetReference()))
				msg.Extensions = extensions
			}
			return base.CreateSchemaProxy(msg)
		}

		s := ScalarFieldToSchema(opts, parent, tt, false)
		return base.CreateSchemaProxy(s)
	}
}

func ScalarFieldToSchema(opts options.Options, parent *base.SchemaProxy, tt protoreflect.FieldDescriptor, inContainer bool) *base.Schema {
	s := &base.Schema{
		ParentProxy: parent,
		Deprecated:  util.IsFieldDeprecated(tt),
	}
	if !inContainer {
		s.Title = string(tt.Name())
		s.Description = util.TypeFieldDescription(opts, tt)
	}

	switch tt.Kind() {
	case protoreflect.BoolKind:
		s.Type = []string{"boolean"}
	case protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sint32Kind: // int32 types
		s.Type = []string{"integer"}
		s.Format = "int32"
	case protoreflect.Fixed32Kind, protoreflect.Uint32Kind: // uint32 types
		s.Type = []string{"integer"}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind: // int64 types
		// NOTE: 64-bit integer types can be strings or numbers because they sometimes
		//       cannot fit into a JSON number type
		s.Type = []string{"integer", "string"}
		s.Format = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind: // uint64 types
		s.Type = []string{"integer", "string"}
		s.Format = "int64"
	case protoreflect.DoubleKind:
		s.Type = []string{"number"}
		s.Format = "double"
	case protoreflect.FloatKind:
		s.Type = []string{"number"}
		s.Format = "float"
	case protoreflect.StringKind:
		s.Type = []string{"string"}
	case protoreflect.BytesKind:
		s.Type = []string{"string"}
		s.Format = "byte"
	}
	// Apply Updates from Options
	s = opts.FieldAnnotator.AnnotateField(opts, s, tt, inContainer)
	return s
}

func ReferenceFieldToSchema(opts options.Options, parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.SchemaProxy {
	switch tt.Kind() {
	case protoreflect.MessageKind:
		opts.FieldReferenceAnnotator.AnnotateFieldReference(opts, parent.Schema(), tt)
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Message().FullName()))
	case protoreflect.EnumKind:
		opts.FieldReferenceAnnotator.AnnotateFieldReference(opts, parent.Schema(), tt)
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Enum().FullName()))
	default:
		panic(fmt.Errorf("ReferenceFieldToSchema called with unknown kind: %T", tt.Kind()))
	}
}

func makeOneOfGroup(opts options.Options, fields []protoreflect.FieldDescriptor) *base.SchemaProxy {
	rootSchemas := make([]*base.SchemaProxy, 0, len(fields))
	for _, field := range fields {
		schema := &base.Schema{
			/*
				Having the title prop here is redundant because it'll also be included inside the
				properties object. But some platforms like readme.io require it to be at the top level for them to
				properly interpret it.
			*/
			Title:      string(field.Name()),
			Properties: orderedmap.New[string, *base.SchemaProxy](),
		}

		fieldName := util.MakeFieldName(opts, field)
		propSchema := FieldToSchema(opts, base.CreateSchemaProxy(schema), field)
		schema.Properties.Set(fieldName, propSchema)
		schema.Required = []string{fieldName}

		rootSchemas = append(rootSchemas, base.CreateSchemaProxy(schema))
	}

	return base.CreateSchemaProxy(&base.Schema{OneOf: rootSchemas})
}

func appendType(s *base.Schema, newType string) {
	if s.Type == nil {
		s.Type = []string{newType}
		return
	}
	if !slices.Contains(s.Type, newType) {
		s.Type = append(s.Type, newType)
	}
}
