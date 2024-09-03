package schema

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageToSchema(tt protoreflect.MessageDescriptor) (string, *base.Schema) {
	slog.Debug("messageToSchema", slog.Any("descriptor", tt.FullName()))
	defer slog.Debug("/messageToSchema", slog.Any("descriptor", tt.FullName()))
	if util.IsWellKnown(tt) {
		wk := util.WellKnownToSchema(tt)
		if wk == nil {
			return "", nil
		}
		return wk.ID, wk.Schema
	}
	s := &base.Schema{
		Title:                string(tt.Name()),
		Description:          util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
		Type:                 []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
	}

	oneOneGroups := map[protoreflect.FullName][]string{}

	props := orderedmap.New[string, *base.SchemaProxy]()
	fields := tt.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if oneOf := field.ContainingOneof(); oneOf != nil {
			oneOneGroups[oneOf.FullName()] = append(oneOneGroups[oneOf.FullName()], field.JSONName())
		}
		props.Set(field.JSONName(), FieldToSchema(base.CreateSchemaProxy(s), field))
	}

	s.Properties = props

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
			slices.Sort(items)
			allOfs = append(allOfs, makeOneOfGroup(items))
		}
		if len(allOfs) == 1 {
			s.AnyOf = allOfs[0].Schema().AnyOf
		}
		s.AllOf = append(s.AllOf, allOfs...)
	}

	// Apply Updates from Options
	s = protovalidate.SchemaWithMessageAnnotations(s, tt)
	s = gnostic.SchemaWithSchemaAnnotations(s, tt)
	return string(tt.FullName()), s
}

func FieldToSchema(parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.SchemaProxy {
	slog.Debug("FieldToSchema", slog.Any("descriptor", tt.FullName()))
	defer slog.Debug("/FieldToSchema", slog.Any("descriptor", tt.FullName()))

	if tt.IsMap() {
		// Handle maps
		root := ScalarFieldToSchema(parent, tt)
		root.Title = string(tt.Name())
		root.Type = []string{"object"}
		root.Description = util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt))
		return base.CreateSchemaProxy(root)
	} else if tt.IsList() {
		var itemSchema *base.SchemaProxy
		switch tt.Kind() {
		case protoreflect.MessageKind:
			itemSchema = ReferenceFieldToSchema(parent, tt)
		case protoreflect.EnumKind:
			itemSchema = ReferenceFieldToSchema(parent, tt)
		default:
			itemSchema = base.CreateSchemaProxy(ScalarFieldToSchema(parent, tt))
		}

		s := &base.Schema{
			Type:  []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{A: itemSchema},
		}
		return base.CreateSchemaProxy(s)
	} else {
		switch tt.Kind() {
		case protoreflect.MessageKind:
			return ReferenceFieldToSchema(parent, tt)
		case protoreflect.EnumKind:
			return ReferenceFieldToSchema(parent, tt)
		}

		return base.CreateSchemaProxy(ScalarFieldToSchema(parent, tt))
	}
}

func ScalarFieldToSchema(parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.Schema {
	s := &base.Schema{
		ParentProxy:          parent,
		Title:                string(tt.Name()),
		Description:          util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
	}

	if tt.IsMap() {
		s.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{A: FieldToSchema(parent, tt.MapValue())}
	}

	switch tt.Kind() {
	case protoreflect.BoolKind:
		s.Type = []string{"boolean"}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		s.Type = []string{"integer"}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		// NOTE: 64-bit types can be strings or numbers because they sometimes
		//       cannot fit into a JSON number type
		s.OneOf = []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
			base.CreateSchemaProxy(&base.Schema{Type: []string{"number"}}),
		}
	case protoreflect.FloatKind:
		s.Type = []string{"number"}
	case protoreflect.StringKind:
		s.Type = []string{"string"}
	case protoreflect.BytesKind:
		s.Type = []string{"string"}
		s.Format = "byte"
	}
	// Apply Updates from Options
	s = protovalidate.SchemaWithFieldAnnotations(s, tt)
	s = gnostic.SchemaWithPropertyAnnotations(s, tt)
	return s
}

func ReferenceFieldToSchema(parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.SchemaProxy {
	switch tt.Kind() {
	case protoreflect.MessageKind:
		protovalidate.PopulateParentProperties(parent.Schema(), tt)
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Message().FullName()))
	case protoreflect.EnumKind:
		protovalidate.PopulateParentProperties(parent.Schema(), tt)
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Enum().FullName()))
	default:
		panic(fmt.Errorf("ReferenceFieldToSchema called with unknown kind: %T", tt.Kind()))
	}
}

func makeOneOfGroup(fields []string) *base.SchemaProxy {
	nestedSchemas := make([]*base.SchemaProxy, 0, len(fields))
	rootSchemas := make([]*base.SchemaProxy, 0, len(fields)+1)
	for _, field := range fields {
		rootSchemas = append(rootSchemas, base.CreateSchemaProxy(&base.Schema{Required: []string{field}}))
		nestedSchemas = append(nestedSchemas, base.CreateSchemaProxy(&base.Schema{Required: []string{field}}))
	}

	rootSchemas = append(rootSchemas, base.CreateSchemaProxy(&base.Schema{Not: base.CreateSchemaProxy(&base.Schema{AnyOf: nestedSchemas})}))
	return base.CreateSchemaProxy(&base.Schema{AnyOf: rootSchemas})
}
