package converter

import (
	"fmt"
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

func TestGeneratorWithOptions(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		generator, err := generatorWithOptions()
		require.NoError(t, err)
		assert.Equal(t, options.NewOptions(), generator.options)
	})
	t.Run("every option", func(t *testing.T) {
		files := new(protoregistry.Files)
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		generator, err := generatorWithOptions(
			WithFiles(files),
			WithFormat("json"),
			WithBaseOpenAPI([]byte("hello!")),
			WithAllowGET(true),
			WithContentTypes("connect+json"),
			WithIncludeNumberEnumValues(true),
			WithStreaming(true),
			WithDebug(true),
			WithProtoAnnotations(true),
		)
		require.NoError(t, err)

		assert.Equal(t, "json", generator.options.Format)
		assert.Equal(t, []byte("hello!"), generator.options.BaseOpenAPI)
		assert.Equal(t, true, generator.options.AllowGET)
		assert.Equal(t, map[string]struct{}{"connect+json": {}}, generator.options.ContentTypes)
		assert.Equal(t, true, generator.options.IncludeNumberEnumValues)
		assert.Equal(t, true, generator.options.WithStreaming)
		assert.Equal(t, true, generator.options.Debug)
		assert.Equal(t, true, generator.options.WithProtoAnnotations)
		assert.Equal(t, []string{"connectrpc/eliza/v1/eliza.proto"}, generator.req.FileToGenerate)
		assert.Equal(
			t,
			[]*descriptorpb.FileDescriptorProto{protodesc.ToFileDescriptorProto(elizav1.File_connectrpc_eliza_v1_eliza_proto)},
			generator.req.ProtoFile)
	})
}

func TestGenerateSingle(t *testing.T) {
	b, err := GenerateSingle(WithGlobal())
	require.NoError(t, err)
	assert.Greater(t, len(b), 4000)
}

func TestGenerate(t *testing.T) {
	files := new(protoregistry.Files)
	require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
	outFiles, err := Generate(
		WithFiles(files),
		WithDebug(true),
	)
	require.NoError(t, err)
	fmt.Println(outFiles)
	require.Len(t, outFiles, 1)
	assert.Greater(t, len(*outFiles[0].Content), 4000)
}
