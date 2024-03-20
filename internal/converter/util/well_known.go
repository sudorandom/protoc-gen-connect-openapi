package util

import (
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var wellKnownToSchemaFns = map[string]func(protoreflect.MessageDescriptor) *jsonschema.Schema{
	"google.protobuf.Duration":  googleDuration,
	"google.protobuf.Timestamp": googleTimestamp,
	"google.protobuf.Value":     googleValue,
	"google.protobuf.Empty":     googleEmpty,
}

func IsWellKnown(msg protoreflect.MessageDescriptor) bool {
	_, ok := wellKnownToSchemaFns[string(msg.FullName())]
	return ok
}

func wellKnownToSchema(msg protoreflect.MessageDescriptor) *jsonschema.Schema {
	fn, ok := wellKnownToSchemaFns[string(msg.FullName())]
	if !ok {
		return nil
	}
	return fn(msg)
}

func googleDuration(msg protoreflect.MessageDescriptor) *jsonschema.Schema {
	s := &jsonschema.Schema{}
	s.WithID(string(msg.FullName()))
	s.WithDescription(FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)))
	s.WithType(jsonschema.String.Type())
	s.WithFormat("regex")
	s.WithPattern(`^[-\+]?([0-9]+\.?[0-9]*|\.[0-9]+)s$`)
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})
	return s
}

func googleTimestamp(msg protoreflect.MessageDescriptor) *jsonschema.Schema {
	s := &jsonschema.Schema{}
	s.WithID(string(msg.FullName()))
	s.WithDescription(FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)))
	s.WithType(jsonschema.String.Type())
	s.WithFormat("date-time")
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})
	return s
}

func googleValue(msg protoreflect.MessageDescriptor) *jsonschema.Schema {
	s := &jsonschema.Schema{}
	s.WithID(string(msg.FullName()))
	s.WithDescription(FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)))
	s.OneOf = []jsonschema.SchemaOrBool{
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Null.Type())},
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Number.Type())},
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.String.Type())},
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Boolean.Type())},
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Array.Type())},
		{TypeObject: (&jsonschema.Schema{}).WithType(jsonschema.Object.Type()).WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(true)})},
	}
	s.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})
	return s
}

func googleEmpty(msg protoreflect.MessageDescriptor) *jsonschema.Schema {
	return nil
}

func IsEmpty(msg protoreflect.MessageDescriptor) bool {
	return msg.FullName() == "google.protobuf.Empty"
}
