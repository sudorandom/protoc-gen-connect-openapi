package converter

import (
	"testing"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestFormatChannelPath(t *testing.T) {
	tests := []struct {
		name     string
		template string
		pkg      string
		service  string
		method   string
		expected string
	}{
		{
			name:     "with package",
			template: "/ws/{package}.{service}/{method}",
			pkg:      "foo.v1",
			service:  "ChatService",
			method:   "ChatStream",
			expected: "/ws/foo.v1.ChatService/ChatStream",
		},
		{
			name:     "without package",
			template: "/ws/{package}.{service}/{method}",
			pkg:      "",
			service:  "ChatService",
			method:   "ChatStream",
			expected: "/ws/ChatService/ChatStream",
		},
		{
			name:     "custom template",
			template: "/stream/{service}/{method}",
			pkg:      "foo.v1",
			service:  "ChatService",
			method:   "ChatStream",
			expected: "/stream/ChatService/ChatStream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatChannelPath(tt.template, tt.pkg, tt.service, tt.method)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateAsyncAPI(t *testing.T) {
	t.Run("no streaming methods returns empty", func(t *testing.T) {
		fdProto := &descriptorpb.FileDescriptorProto{
			Name:    proto.String("unary.proto"),
			Package: proto.String("test.unary"),
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("Req")},
				{Name: proto.String("Resp")},
			},
			Service: []*descriptorpb.ServiceDescriptorProto{
				{
					Name: proto.String("UnaryService"),
					Method: []*descriptorpb.MethodDescriptorProto{
						{
							Name:       proto.String("UnaryCall"),
							InputType:  proto.String(".test.unary.Req"),
							OutputType: proto.String(".test.unary.Resp"),
						},
					},
				},
			},
		}

		fd, err := protodesc.NewFile(fdProto, nil)
		require.NoError(t, err)

		opts := options.NewOptions()
		out, err := GenerateAsyncAPI(opts, []protoreflect.FileDescriptor{fd})
		require.NoError(t, err)
		assert.Empty(t, out)
	})
}

func TestGenerateAsyncAPICollidingNames(t *testing.T) {
	// Two packages with the same service and method names must both keep
	// their channels and operations.
	streamingFile := func(pkg string) protoreflect.FileDescriptor {
		fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
			Name:    proto.String(pkg + ".proto"),
			Package: proto.String(pkg),
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("Msg")},
			},
			Service: []*descriptorpb.ServiceDescriptorProto{{
				Name: proto.String("S"),
				Method: []*descriptorpb.MethodDescriptorProto{{
					Name:            proto.String("Stream"),
					InputType:       proto.String("." + pkg + ".Msg"),
					OutputType:      proto.String("." + pkg + ".Msg"),
					ClientStreaming: proto.Bool(true),
					ServerStreaming: proto.Bool(true),
				}},
			}},
		}, nil)
		require.NoError(t, err)
		return fd
	}

	opts := options.NewOptions()
	opts.MessageAnnotator = &annotator{}
	opts.FieldAnnotator = &annotator{}
	opts.FieldReferenceAnnotator = &annotator{}
	out, err := GenerateAsyncAPI(opts, []protoreflect.FileDescriptor{streamingFile("a"), streamingFile("b")})
	require.NoError(t, err)
	assert.Contains(t, out, "SStreamChannel:")
	assert.Contains(t, out, "b_SStreamChannel:")
	assert.Contains(t, out, "/ws/a.S/Stream")
	assert.Contains(t, out, "/ws/b.S/Stream")
	assert.Contains(t, out, "title: S, b.S WebSockets")
}
