package util

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var wellKnownToSchemaFns = map[string]func(protoreflect.MessageDescriptor) *IDSchema{
	"google.protobuf.Duration":  googleDuration,
	"google.protobuf.Timestamp": googleTimestamp,
	"google.protobuf.Value":     googleValue,
	"google.protobuf.Empty":     googleEmpty,
	"google.protobuf.Any":       func(_ protoreflect.MessageDescriptor) *IDSchema { return NewGoogleAny() },
}

type IDSchema struct {
	ID     string
	Schema *base.Schema
}

func IsWellKnown(msg protoreflect.MessageDescriptor) bool {
	_, ok := wellKnownToSchemaFns[string(msg.FullName())]
	return ok
}

func wellKnownToSchema(msg protoreflect.MessageDescriptor) *IDSchema {
	fn, ok := wellKnownToSchemaFns[string(msg.FullName())]
	if !ok {
		return nil
	}
	return fn(msg)
}

func googleDuration(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description:          FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:                 []string{"string"},
			Format:               "regex",
			Pattern:              `^[-\+]?([0-9]+\.?[0-9]*|\.[0-9]+)s$`,
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
		},
	}
}

func googleTimestamp(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description:          FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:                 []string{"string"},
			Format:               "date-time",
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false},
		},
	}
}

func googleValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			OneOf: []*base.SchemaProxy{
				base.CreateSchemaProxy(&base.Schema{Type: []string{"null"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"number"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"boolean"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"array"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"object"}, AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: false}}),
			},
		},
	}
}

func googleEmpty(msg protoreflect.MessageDescriptor) *IDSchema {
	return nil
}

func NewGoogleAny() *IDSchema {
	typeS := &base.Schema{
		Type:                 []string{"string"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
	}
	props := orderedmap.New[string, *base.SchemaProxy]()
	props.Set("@type", base.CreateSchemaProxy(typeS))

	return &IDSchema{
		ID: "google.protobuf.Any",
		Schema: &base.Schema{
			Description:          "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.",
			Type:                 []string{"object"},
			Properties:           props,
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
		},
	}
}

func IsEmpty(msg protoreflect.MessageDescriptor) bool {
	return msg.FullName() == "google.protobuf.Empty"
}
