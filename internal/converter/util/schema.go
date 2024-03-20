package util

import (
	"log/slog"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageToSchema(tt protoreflect.MessageDescriptor) *jsonschema.Schema {
	slog.Debug("messageToSchema", slog.Any("descriptor", tt.FullName()))
	if IsWellKnown(tt) {
		return wellKnownToSchema(tt)
	}
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithType(jsonschema.Object.Type())
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})

	fields := tt.Fields()
	children := make(map[string]jsonschema.SchemaOrBool, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		child := FieldToSchema(s, field)
		children[field.JSONName()] = jsonschema.SchemaOrBool{TypeObject: child}
	}

	s.WithProperties(children)

	// Apply Updates from Options
	s = protovalidate.SchemaWithMessageAnnotations(s, tt)
	s = gnostic.SchemaWithSchemaAnnotations(s, tt)
	return s
}

func FieldToSchema(parent *jsonschema.Schema, tt protoreflect.FieldDescriptor) *jsonschema.Schema {
	slog.Debug("FieldToSchema", slog.Any("descriptor", tt.FullName()))
	s := &jsonschema.Schema{Parent: parent}

	// TODO: 64-bit types can be strings or numbers because they sometimes
	//       cannot fit into a JSON number type
	switch tt.Kind() {
	case protoreflect.BoolKind:
		s.WithType(jsonschema.Boolean.Type())
	case protoreflect.EnumKind:
		s.WithRef("#/components/schemas/" + string(tt.Enum().FullName()))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		s.WithType(jsonschema.Integer.Type())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		s.WithOneOf(
			jsonschema.SchemaOrBool{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.String.Type())},
			jsonschema.SchemaOrBool{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Number.Type())},
		)
	case protoreflect.FloatKind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.StringKind:
		s.WithType(jsonschema.String.Type())
	case protoreflect.BytesKind:
		s.WithType(jsonschema.String.Type())
		s.WithFormat("byte")
	case protoreflect.MessageKind:
		s.WithRef("#/components/schemas/" + string(tt.Message().FullName()))
		s.WithType(jsonschema.Object.Type())
	}

	// Handle maps
	if tt.IsMap() {
		s.AdditionalProperties = &jsonschema.SchemaOrBool{TypeObject: FieldToSchema(s, tt.MapValue())}
		s.WithType(jsonschema.Object.Type())
		s.Ref = nil
	}

	// Handle Lists
	if tt.IsList() {
		wrapped := s
		s = &jsonschema.Schema{}
		wrapped.Parent = s
		s.WithType(jsonschema.Array.Type())
		s.WithItems(jsonschema.Items{SchemaOrBool: &jsonschema.SchemaOrBool{TypeObject: wrapped}})
	}

	s.WithTitle(string(tt.Name()))
	s.WithDescription(FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})

	// Apply Updates from Options
	s = protovalidate.SchemaWithFieldAnnotations(s, tt)
	s = gnostic.SchemaWithPropertyAnnotations(s, tt)
	return s
}
