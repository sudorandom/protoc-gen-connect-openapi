package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.True(t, IsWellKnown(tt.md))
			res := WellKnownToSchema(tt.md)
			require.NotNil(t, res)
			tt.check(t, res)
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
		assert.Nil(t, WellKnownToSchema(md))
	})
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty((&emptypb.Empty{}).ProtoReflect().Descriptor()))
	assert.False(t, IsEmpty((&durationpb.Duration{}).ProtoReflect().Descriptor()))
}
