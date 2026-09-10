package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestWellKnownToSchema(t *testing.T) {
	tests := []struct {
		name    string
		md      protoreflect.MessageDescriptor
		isKnown bool
		check   func(t *testing.T, idSchema *IDSchema)
	}{
		{
			name:    "google.protobuf.Duration",
			md:      (&durationpb.Duration{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Duration", s.ID)
				assert.Equal(t, []string{"string"}, s.Schema.Type)
				assert.Equal(t, "duration", s.Schema.Format)
			},
		},
		{
			name:    "google.protobuf.Timestamp",
			md:      (&timestamppb.Timestamp{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Timestamp", s.ID)
				assert.Equal(t, []string{"string"}, s.Schema.Type)
				assert.Equal(t, "date-time", s.Schema.Format)
				assert.NotEmpty(t, s.Schema.Examples)
			},
		},
		{
			name:    "google.protobuf.Empty",
			md:      (&emptypb.Empty{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Empty", s.ID)
				assert.Equal(t, []string{"object"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.Any",
			md:      (&anypb.Any{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Any", s.ID)
				assert.Equal(t, []string{"object"}, s.Schema.Type)
				assert.NotNil(t, s.Schema.Properties)
				assert.NotNil(t, s.Schema.AdditionalProperties)
			},
		},
		{
			name:    "google.protobuf.FieldMask",
			md:      (&fieldmaskpb.FieldMask{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.FieldMask", s.ID)
				assert.Equal(t, []string{"string"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.Struct",
			md:      (&structpb.Struct{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Struct", s.ID)
				assert.Equal(t, []string{"object"}, s.Schema.Type)
				assert.NotNil(t, s.Schema.AdditionalProperties)
			},
		},
		{
			name:    "google.protobuf.Value",
			md:      (&structpb.Value{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Value", s.ID)
				assert.NotEmpty(t, s.Schema.OneOf)
			},
		},
		{
			name:    "google.protobuf.ListValue",
			md:      (&structpb.ListValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.ListValue", s.ID)
				// ProtoJSON encodes ListValue as a bare array, not an object with a `values` field.
				assert.Equal(t, []string{"array"}, s.Schema.Type)
				assert.Nil(t, s.Schema.Properties)
				require.NotNil(t, s.Schema.Items)
				require.NotNil(t, s.Schema.Items.A)
				assert.Equal(t, "#/components/schemas/google.protobuf.Value", s.Schema.Items.A.GetReference())
			},
		},
		{
			name:    "google.protobuf.StringValue",
			md:      (&wrapperspb.StringValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.StringValue", s.ID)
				assert.Equal(t, []string{"string"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.BytesValue",
			md:      (&wrapperspb.BytesValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.BytesValue", s.ID)
				assert.Equal(t, []string{"string"}, s.Schema.Type)
				assert.Equal(t, "binary", s.Schema.Format)
			},
		},
		{
			name:    "google.protobuf.BoolValue",
			md:      (&wrapperspb.BoolValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.BoolValue", s.ID)
				assert.Equal(t, []string{"boolean"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.DoubleValue",
			md:      (&wrapperspb.DoubleValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.DoubleValue", s.ID)
				assert.NotEmpty(t, s.Schema.OneOf)
			},
		},
		{
			name:    "google.protobuf.Int32Value",
			md:      (&wrapperspb.Int32Value{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Int32Value", s.ID)
				assert.Equal(t, []string{"number"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.Uint32Value",
			md:      (&wrapperspb.UInt32Value{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.UInt32Value", s.ID)
				assert.Equal(t, []string{"number"}, s.Schema.Type)
			},
		},
		{
			name:    "google.protobuf.Int64Value",
			md:      (&wrapperspb.Int64Value{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.Int64Value", s.ID)
				assert.NotEmpty(t, s.Schema.OneOf)
			},
		},
		{
			name:    "google.protobuf.Uint64Value",
			md:      (&wrapperspb.UInt64Value{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.UInt64Value", s.ID)
				assert.NotEmpty(t, s.Schema.OneOf)
			},
		},
		{
			name:    "google.protobuf.FloatValue",
			md:      (&wrapperspb.FloatValue{}).ProtoReflect().Descriptor(),
			isKnown: true,
			check: func(t *testing.T, s *IDSchema) {
				assert.Equal(t, "google.protobuf.FloatValue", s.ID)
				assert.NotEmpty(t, s.Schema.OneOf)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, ok := expectedConciseWellKnownDescriptions[string(tt.md.FullName())]
			require.True(t, ok, "missing expected concise description for %s", tt.md.FullName())

			res := WellKnownToSchema(options.NewOptions(), tt.md)
			require.NotNil(t, res)
			assert.Equal(t, want, res.Schema.Description)
			tt.check(t, res)

			omitOpts := options.NewOptions()
			omitOpts.WellKnownTypeDescriptions = options.WellKnownTypeDescriptionsOmit
			withoutDescription := WellKnownToSchema(omitOpts, tt.md)
			require.NotNil(t, withoutDescription)
			assert.Empty(t, withoutDescription.Schema.Description)
			tt.check(t, withoutDescription)

			conciseOpts := options.NewOptions()
			conciseOpts.WellKnownTypeDescriptions = options.WellKnownTypeDescriptionsConcise
			concise := WellKnownToSchema(conciseOpts, tt.md)
			require.NotNil(t, concise)
			assert.Equal(t, want, concise.Schema.Description)
			assert.NotContains(t, concise.Schema.Description, "\n")
			tt.check(t, concise)

			fullOpts := options.NewOptions()
			fullOpts.WellKnownTypeDescriptions = options.WellKnownTypeDescriptionsFull
			full := WellKnownToSchema(fullOpts, tt.md)
			require.NotNil(t, full)
			tt.check(t, full)
		})
	}

	t.Run("unknown message", func(t *testing.T) {
		fdProto := &descriptorpb.FileDescriptorProto{
			Name:    proto.String("custom.proto"),
			Package: proto.String("custom"),
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("CustomMessage")},
			},
		}
		fd, err := protodesc.NewFile(fdProto, nil)
		require.NoError(t, err)
		md := fd.Messages().Get(0)

		assert.False(t, IsWellKnown(md))
		assert.Nil(t, WellKnownToSchema(options.NewOptions(), md))
	})
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty((&emptypb.Empty{}).ProtoReflect().Descriptor()))
	assert.False(t, IsEmpty((&durationpb.Duration{}).ProtoReflect().Descriptor()))
}

// expectedConciseWellKnownDescriptions are JSON-oriented one-liners for well-known types.
var expectedConciseWellKnownDescriptions = map[string]string{
	"google.protobuf.Timestamp":   "A point in time in RFC 3339 format, with up to nanosecond precision. Output uses UTC (`Z`); input may use an offset from UTC.",
	"google.protobuf.Duration":    "A signed duration in seconds, with up to nine fractional digits and an `s` suffix (for example, `3s` or `-0.001s`).",
	"google.protobuf.Empty":       "An empty JSON object.",
	"google.protobuf.Any":         "An arbitrary message with a type URL identifying the encoded message type.",
	"google.protobuf.FieldMask":   "Comma-separated field paths in lowerCamelCase (for example, `user.displayName,photo`).",
	"google.protobuf.Struct":      "A JSON object whose property values may be any JSON value.",
	"google.protobuf.Value":       "Any JSON value: `null`, number, string, boolean, array, or object.",
	"google.protobuf.ListValue":   "A JSON array whose elements may be any JSON value.",
	"google.protobuf.NullValue":   "The JSON `null` value.",
	"google.protobuf.StringValue": "A string value; `null` is also accepted.",
	"google.protobuf.BytesValue":  "A base64-encoded byte string; `null` is also accepted.",
	"google.protobuf.BoolValue":   "A boolean value; `null` is also accepted.",
	"google.protobuf.DoubleValue": "A double-precision number. The non-finite values `NaN`, `Infinity`, and `-Infinity` are encoded as strings; `null` is also accepted.",
	"google.protobuf.FloatValue":  "A single-precision number. The non-finite values `NaN`, `Infinity`, and `-Infinity` are encoded as strings; `null` is also accepted.",
	"google.protobuf.Int64Value":  "A signed 64-bit integer encoded as a decimal string. Input may also be a JSON number or `null`.",
	"google.protobuf.UInt64Value": "An unsigned 64-bit integer encoded as a decimal string. Input may also be a JSON number or `null`.",
	"google.protobuf.Int32Value":  "A signed 32-bit integer; `null` is also accepted.",
	"google.protobuf.UInt32Value": "An unsigned 32-bit integer; `null` is also accepted.",
}
