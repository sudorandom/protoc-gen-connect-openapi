package util

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

var wellKnownToSchemaFns = map[string]func(protoreflect.MessageDescriptor) *IDSchema{
	"google.protobuf.Duration":  googleDuration,
	"google.protobuf.Timestamp": googleTimestamp,
	"google.protobuf.Empty":     googleEmpty,
	"google.protobuf.Any":       func(_ protoreflect.MessageDescriptor) *IDSchema { return NewGoogleAny() },
	"google.protobuf.FieldMask": googleFieldmask,

	// google.protobuf.[Type]Value
	"google.protobuf.Struct":      googleStruct,
	"google.protobuf.Value":       googleValue,
	"google.protobuf.NullValue":   googleNullValue,
	"google.protobuf.StringValue": googleStringValue,
	"google.protobuf.BytesValue":  googleBytesValue,
	"google.protobuf.BoolValue":   googleBoolValue,
	"google.protobuf.DoubleValue": google64BitNumberValue,
	"google.protobuf.Int64Value":  google64BitNumberValue,
	"google.protobuf.Uint64Value": google64BitNumberValue,
	"google.protobuf.FloatValue":  google64BitNumberValue,
	"google.protobuf.Int32Value":  google32BitNumberValue,
	"google.protobuf.Uint32Value": google32BitNumberValue,
}

type IDSchema struct {
	ID     string
	Schema *base.Schema
}

func IsWellKnown(msg protoreflect.MessageDescriptor) bool {
	_, ok := wellKnownToSchemaFns[string(msg.FullName())]
	return ok
}

func WellKnownToSchema(msg protoreflect.MessageDescriptor) *IDSchema {
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
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
			Format:      "duration",
		},
	}
}

func googleTimestamp(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
			Format:      "date-time",
			Examples: []*yaml.Node{
				utils.CreateStringNode("1s"),
				utils.CreateStringNode("1.000340012s"),
			},
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
				base.CreateSchemaProxy(&base.Schema{
					Type:                 []string{"object"},
					AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
				}),
			},
		},
	}
}

func googleStruct(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				A: base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Value"),
			},
		},
	}
}

func googleNullValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"null"},
		},
	}
}

func googleStringValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
		},
	}
}

func googleBoolValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"boolean"},
		},
	}
}

func googleBytesValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
			Format:      "binary",
		},
	}
}

func google32BitNumberValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"number"},
		},
	}
}

func google64BitNumberValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			OneOf: []*base.SchemaProxy{
				base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
				base.CreateSchemaProxy(&base.Schema{Type: []string{"number"}}),
			},
		},
	}
}

func googleEmpty(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"object"},
		},
	}
}

func NewGoogleAny() *IDSchema {
	props := orderedmap.New[string, *base.SchemaProxy]()
	props.Set("type", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
	props.Set("value", base.CreateSchemaProxy(&base.Schema{
		Type:   []string{"string"},
		Format: "binary",
	}))
	props.Set("debug", base.CreateSchemaProxy(&base.Schema{
		Type:                 []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
	}))

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

func googleFieldmask(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
		},
	}
}

func IsEmpty(msg protoreflect.MessageDescriptor) bool {
	return msg.FullName() == "google.protobuf.Empty"
}
