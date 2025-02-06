package schema

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageToSchema(opts options.Options, tt protoreflect.MessageDescriptor) (string, *base.Schema) {
	slog.Debug("messageToSchema", slog.Any("descriptor", tt.FullName()))
	defer slog.Debug("/messageToSchema", slog.Any("descriptor", tt.FullName()))
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

	oneOneGroups := map[protoreflect.FullName][]string{}

	props := orderedmap.New[string, *base.SchemaProxy]()
	fields := tt.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if oneOf := field.ContainingOneof(); oneOf != nil && !oneOf.IsSynthetic() {
			oneOneGroups[oneOf.FullName()] = append(oneOneGroups[oneOf.FullName()], util.MakeFieldName(opts, field))
		}
		prop := FieldToSchema(opts, base.CreateSchemaProxy(s), field)
		if field.HasOptionalKeyword() {
			nullable := true
			prop.Schema().Nullable = &nullable
		}
		props.Set(util.MakeFieldName(opts, field), prop)
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
		} else {
			s.AllOf = append(s.AllOf, allOfs...)
		}
	}

	// Apply Updates from Options
	s = opts.MessageAnnotator.AnnotateMessage(opts, s, tt)
	return string(tt.FullName()), s
}

func FieldToSchema(opts options.Options, parent *base.SchemaProxy, tt protoreflect.FieldDescriptor) *base.SchemaProxy {
	slog.Debug("FieldToSchema", slog.Any("descriptor", tt.FullName()))
	defer slog.Debug("/FieldToSchema", slog.Any("descriptor", tt.FullName()))

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
		}
		s = opts.FieldAnnotator.AnnotateField(opts, s, tt, false)
		return base.CreateSchemaProxy(s)
	} else {
		switch tt.Kind() {
		case protoreflect.MessageKind, protoreflect.EnumKind:
			msg := ScalarFieldToSchema(opts, parent, tt, false)
			msg.AllOf = append(msg.AllOf, ReferenceFieldToSchema(opts, parent, tt))
			return base.CreateSchemaProxy(msg)
		}

		return base.CreateSchemaProxy(ScalarFieldToSchema(opts, parent, tt, false))
	}
}

func ScalarFieldToSchema(opts options.Options, parent *base.SchemaProxy, tt protoreflect.FieldDescriptor, inContainer bool) *base.Schema {
	s := &base.Schema{
		ParentProxy: parent,
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
	s = opts.FieldAnnotator.AnnotateField(opts, s, tt, false)
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
