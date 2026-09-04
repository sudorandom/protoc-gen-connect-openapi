package converter

import (
	"testing"

	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
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
			WithOnlyGoogleapiHTTP(true),
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
		assert.Equal(t, true, generator.options.OnlyGoogleapiHTTP)
		assert.Equal(t, []string{"connectrpc/eliza/v1/eliza.proto"}, generator.req.FileToGenerate)
		assert.Equal(
			t,
			[]*descriptorpb.FileDescriptorProto{protodesc.ToFileDescriptorProto(elizav1.File_connectrpc_eliza_v1_eliza_proto)},
			generator.req.ProtoFile)
	})

	t.Run("all remaining options", func(t *testing.T) {
		files := new(protoregistry.Files)
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		generator, err := generatorWithOptions(
			WithSourceFiles(files),
			WithOverrideOpenAPI([]byte("override")),
			WithIgnoreGoogleapiHTTP(true),
			WithProtoNames(true),
			WithTrimUnusedTypes(true),
			WithServices([]protoreflect.FullName{"connectrpc.eliza.v1.ElizaService"}),
			WithServicePatterns([]string{"connectrpc.eliza.v1.*"}),
			WithShortServiceTags(true),
			WithShortOperationIds(true),
			WithoutDefaultTags(true),
			WithoutFieldBehaviorPrefixes(true),
			WithDisableDefaultResponse(true),
			WithAllowedVisibilities("INTERNAL", "PREVIEW"),
			WithFullyQualifiedMessageNames(true),
			WithServiceDescriptions(true),
			WithPathPrefix("/v1"),
			WithAsyncAPIPath("asyncapi.yaml"),
			WithAsyncAPIChannelTemplate("/ws/{service}/{method}"),
			WithGoogleErrorDetail(true),
			WithLogger(nil),
			WithFeatures(options.FeatureConnectRPC, options.FeatureGoogleAPIHTTP),
		)
		require.NoError(t, err)

		assert.Equal(t, []byte("override"), generator.options.OverrideOpenAPI)
		assert.True(t, generator.options.IgnoreGoogleapiHTTP)
		assert.True(t, generator.options.WithProtoNames)
		assert.True(t, generator.options.TrimUnusedTypes)
		assert.Len(t, generator.options.Services, 2)
		assert.True(t, generator.options.ShortServiceTags)
		assert.True(t, generator.options.ShortOperationIds)
		assert.True(t, generator.options.WithoutDefaultTags)
		assert.True(t, generator.options.WithoutFieldBehaviorPrefixes)
		assert.True(t, generator.options.DisableDefaultResponse)
		assert.True(t, generator.options.AllowedVisibilities["INTERNAL"])
		assert.True(t, generator.options.AllowedVisibilities["PREVIEW"])
		assert.True(t, generator.options.FullyQualifiedMessageNames)
		assert.True(t, generator.options.WithServiceDescriptions)
		assert.Equal(t, "/v1", generator.options.PathPrefix)
		assert.Equal(t, "asyncapi.yaml", generator.options.AsyncAPIPath)
		assert.Equal(t, "/ws/{service}/{method}", generator.options.AsyncAPIChannelTemplate)
		assert.True(t, generator.options.WithGoogleErrorDetail)
		assert.True(t, generator.options.FeatureEnabled(options.FeatureConnectRPC))
		assert.True(t, generator.options.FeatureEnabled(options.FeatureGoogleAPIHTTP))
	})

	t.Run("option errors", func(t *testing.T) {
		t.Run("invalid content type", func(t *testing.T) {
			_, err := generatorWithOptions(WithContentTypes("invalid_content_type"))
			require.Error(t, err)
		})

		t.Run("invalid service pattern in WithServices", func(t *testing.T) {
			_, err := generatorWithOptions(WithServices([]protoreflect.FullName{"["}))
			require.Error(t, err)
		})

		t.Run("invalid service pattern in WithServicePatterns", func(t *testing.T) {
			_, err := generatorWithOptions(WithServicePatterns([]string{"["}))
			require.Error(t, err)
		})

		t.Run("invalid feature in WithFeatures", func(t *testing.T) {
			_, err := generatorWithOptions(WithFeatures("nonexistent_feature"))
			require.Error(t, err)
		})
	})
}

func TestGenerateSingle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, err := GenerateSingle(WithGlobal())
		require.NoError(t, err)
		assert.Greater(t, len(b), 4000)
	})

	t.Run("error in options", func(t *testing.T) {
		_, err := GenerateSingle(WithContentTypes("invalid"))
		require.Error(t, err)
	})
}

func TestGenerate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		files := new(protoregistry.Files)
		require.NoError(t, files.RegisterFile(elizav1.File_connectrpc_eliza_v1_eliza_proto))
		outFiles, err := Generate(
			WithFiles(files),
			WithDebug(true),
		)
		require.NoError(t, err)
		require.Len(t, outFiles, 1)
		assert.Greater(t, len(*outFiles[0].Content), 4000)
	})

	t.Run("error in options", func(t *testing.T) {
		_, err := Generate(WithContentTypes("invalid"))
		require.Error(t, err)
	})
}
