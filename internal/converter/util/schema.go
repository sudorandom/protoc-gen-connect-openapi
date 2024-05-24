package util

import (
	"fmt"
	"log/slog"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageToSchema(tt protoreflect.MessageDescriptor) (string, *base.Schema) {
	slog.Debug("messageToSchema", slog.Any("descriptor", tt.FullName()))
	defer slog.Debug("/messageToSchema", slog.Any("descriptor", tt.FullName()))
	if IsWellKnown(tt) {
		wk := wellKnownToSchema(tt)
		if wk == nil {
			return "", nil
		}
		return wk.ID, wk.Schema
	}
	s := &base.Schema{
		Title:                string(tt.Name()),
		Description:          FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
		Type:                 []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
	}

	props := orderedmap.New[string, *base.SchemaProxy]()
	fields := tt.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		props.Set(field.JSONName(), FieldToSchema(base.CreateSchemaProxy(s), field))
	}

	s.Properties = props

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
		root.Description = FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt))
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
		Description:          FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
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
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Message().FullName()))
	case protoreflect.EnumKind:
		return base.CreateSchemaProxyRef("#/components/schemas/" + string(tt.Enum().FullName()))
	default:
		panic(fmt.Errorf("ReferenceFieldToSchema called with unknown kind: %T", tt.Kind()))
	}
}
