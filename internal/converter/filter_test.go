package converter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestServiceFilterFileGeneration(t *testing.T) {
	// Load descriptor set
	f, err := os.ReadFile(filepath.Join("testdata", "fileset.binpb"))
	require.NoError(t, err)

	pf := new(descriptorpb.FileDescriptorSet)
	require.NoError(t, proto.Unmarshal(f, pf))

	// Make Generation Request
	req := new(pluginpb.CodeGeneratorRequest)
	req.ProtoFile = pf.GetFile()

	// Add all files to generate to simulate a full project run
	for _, f := range req.GetProtoFile() {
		req.FileToGenerate = append(req.FileToGenerate, f.GetName())
	}

	t.Run("per-file mode filters output files", func(t *testing.T) {
		// Filter for a service that only exists in with_service_filters/service_filters.proto
		opts, err := options.FromString("services=testing.filters.v1.UserService")
		require.NoError(t, err)

		resp, err := converter.ConvertWithOptions(req, opts)
		require.NoError(t, err)

		// Check which files were generated
		var generatedFiles []string
		for _, f := range resp.File {
			generatedFiles = append(generatedFiles, f.GetName())
		}

		// We expect ONLY the file containing UserService to be generated
		assert.Len(t, generatedFiles, 1)
		assert.Equal(t, "with_service_filters/service_filters.openapi.yaml", generatedFiles[0])
	})

	t.Run("merged mode generates a single file", func(t *testing.T) {
		// Filter for a service and set a single output path
		opts, err := options.FromString("services=testing.filters.v1.UserService,path=openapi.yaml")
		require.NoError(t, err)

		resp, err := converter.ConvertWithOptions(req, opts)
		require.NoError(t, err)

		assert.Len(t, resp.File, 1)
		assert.Equal(t, "openapi.yaml", resp.File[0].GetName())
	})

	t.Run("no matching services generates no files in per-file mode", func(t *testing.T) {
		opts, err := options.FromString("services=nonexistent.Service")
		require.NoError(t, err)

		resp, err := converter.ConvertWithOptions(req, opts)
		require.NoError(t, err)

		assert.Empty(t, resp.File)
	})
}
