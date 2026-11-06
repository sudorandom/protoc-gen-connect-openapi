// Copyright 2023-2026 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	"google.protobuf.ListValue":   googleListValue,
	"google.protobuf.StringValue": googleStringValue,
	"google.protobuf.BytesValue":  googleBytesValue,
	"google.protobuf.BoolValue":   googleBoolValue,
	"google.protobuf.DoubleValue": google64BitNumberValue,
	"google.protobuf.Int64Value":  google64BitNumberValue,
	"google.protobuf.UInt64Value": google64BitNumberValue,
	"google.protobuf.Uint64Value": google64BitNumberValue,
	"google.protobuf.FloatValue":  google64BitNumberValue,
	"google.protobuf.Int32Value":  google32BitNumberValue,
	"google.protobuf.UInt32Value": google32BitNumberValue,
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

func WellKnownToSchema(opts options.Options, msg protoreflect.MessageDescriptor) *IDSchema {
	fn, ok := wellKnownToSchemaFns[string(msg.FullName())]
	if !ok {
		return nil
	}
	idSchema := fn(msg)
	if idSchema != nil && idSchema.Schema != nil {
		idSchema.Schema.Description = wellKnownDescription(opts, msg, idSchema.Schema.Description)
	}
	return idSchema
}

func wellKnownDescription(opts options.Options, msg protoreflect.MessageDescriptor, full string) string {
	switch opts.WellKnownTypeDescriptions {
	case options.WellKnownTypeDescriptionsOmit:
		return ""
	case options.WellKnownTypeDescriptionsFull:
		return full
	case options.WellKnownTypeDescriptionsConcise, "":
		if concise, ok := conciseWellKnownDescriptions[string(msg.FullName())]; ok {
			return concise
		}
		return full
	default:
		return full
	}
}

// conciseWellKnownDescriptions document the JSON representation of well-known types.
var conciseWellKnownDescriptions = map[string]string{
	"google.protobuf.Timestamp":   "A point in time in RFC 3339 format, with up to nanosecond precision. Output uses UTC (`Z`); input may use an offset from UTC.",
	"google.protobuf.Duration":    "A signed duration in seconds, with up to nine fractional digits and an `s` suffix (for example, `3s` or `-0.001s`).",
	"google.protobuf.Empty":       "An empty JSON object.",
	"google.protobuf.Any":         "An arbitrary message with a type URL identifying the encoded message type.",
	"google.protobuf.FieldMask":   "Comma-separated field paths in lowerCamelCase (for example, `user.displayName,photo`).",
	"google.protobuf.Struct":      "A JSON object whose property values may be any JSON value.",
	"google.protobuf.Value":       "Any JSON value: `null`, number, string, boolean, array, or object.",
	"google.protobuf.ListValue":   "A JSON array whose elements may be any JSON value.",
	"google.protobuf.StringValue": "A string value; `null` is also accepted.",
	"google.protobuf.BytesValue":  "A base64-encoded byte string; `null` is also accepted.",
	"google.protobuf.BoolValue":   "A boolean value; `null` is also accepted.",
	"google.protobuf.DoubleValue": "A double-precision number. The non-finite values `NaN`, `Infinity`, and `-Infinity` are encoded as strings; `null` is also accepted.",
	"google.protobuf.FloatValue":  "A single-precision number. The non-finite values `NaN`, `Infinity`, and `-Infinity` are encoded as strings; `null` is also accepted.",
	"google.protobuf.Int64Value":  "A signed 64-bit integer encoded as a decimal string. Input may also be a JSON number or `null`.",
	"google.protobuf.UInt64Value": "An unsigned 64-bit integer encoded as a decimal string. Input may also be a JSON number or `null`.",
	"google.protobuf.Uint64Value": "An unsigned 64-bit integer encoded as a decimal string. Input may also be a JSON number or `null`.",
	"google.protobuf.Int32Value":  "A signed 32-bit integer; `null` is also accepted.",
	"google.protobuf.UInt32Value": "An unsigned 32-bit integer; `null` is also accepted.",
	"google.protobuf.Uint32Value": "An unsigned 32-bit integer; `null` is also accepted.",
}

// DurationPattern matches protojson's Duration form: seconds with up to nine
// fractional digits and an "s" suffix.
const DurationPattern = `^-?[0-9]+(\.[0-9]{1,9})?s$`

func googleDuration(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string"},
			// JSON Schema's "duration" format is ISO 8601 (P1D), which protojson
			// doesn't use, so the pattern describes the actual form instead.
			Pattern: DurationPattern,
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
				utils.CreateStringNode("2023-01-15T01:30:15.01Z"),
				utils.CreateStringNode("2024-12-25T12:00:00Z"),
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

func googleListValue(msg protoreflect.MessageDescriptor) *IDSchema {
	// ProtoJSON encodes ListValue as a bare JSON array, without the `values` field wrapper.
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				A: base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Value"),
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

func googleStringValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string", "null"},
		},
	}
}

func googleBoolValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"boolean", "null"},
		},
	}
}

func googleBytesValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"string", "null"},
			Format:      "byte",
		},
	}
}

func google32BitNumberValue(msg protoreflect.MessageDescriptor) *IDSchema {
	return &IDSchema{
		ID: string(msg.FullName()),
		Schema: &base.Schema{
			Description: FormatComments(msg.ParentFile().SourceLocations().ByDescriptor(msg)),
			Type:        []string{"integer", "null"},
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
				base.CreateSchemaProxy(&base.Schema{Type: []string{"null"}}),
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
	// protojson writes an Any as "@type" next to the embedded message's own
	// fields (or a "value" field for well-known types with special JSON).
	props := orderedmap.New[string, *base.SchemaProxy]()
	props.Set("@type", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A URL identifying the type of the embedded message, e.g. `type.googleapis.com/google.rpc.ErrorInfo`.",
	}))

	return &IDSchema{
		ID: "google.protobuf.Any",
		Schema: &base.Schema{
			Description:          "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.",
			Type:                 []string{"object"},
			Properties:           props,
			Required:             []string{"@type"},
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
