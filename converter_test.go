package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pseudomuto/protokit/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestConvert calls greetings.Hello with a name, checking
// for a valid return value.
func TestConvert(t *testing.T) {
	paths, err := filepath.Glob("fixtures/*.proto")
	require.NoError(t, err)
	for _, protofile := range paths {
		protofile := protofile
		t.Run(protofile, func(t *testing.T) {
			relPath := strings.TrimPrefix(protofile, "fixtures/")
			pf, err := utils.LoadDescriptorSet("fixtures", "fileset.binpb")
			req := utils.CreateGenRequest(pf, relPath)
			require.NoError(t, err)

			b, err := proto.Marshal(req)
			require.NoError(t, err)

			resp, err := ConvertFrom(bytes.NewBuffer(b))
			require.NoError(t, err)
			assert.Len(t, resp.File, 1)
			assert.NotNil(t, resp.File[0].Name)
			assert.Equal(t, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".openapi.json", resp.File[0].GetName())

			expectedFile, err := os.ReadFile(strings.TrimSuffix(protofile, filepath.Ext(protofile)) + ".openapi.json")
			require.NoError(t, err)

			assert.Equal(t, string(expectedFile), resp.File[0].GetContent())
		})
	}
}
