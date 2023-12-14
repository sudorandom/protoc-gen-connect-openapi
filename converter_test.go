package main

import (
	"bytes"
	"fmt"
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
		fmt.Println(protofile)
		t.Run(protofile, func(t *testing.T) {
			pf, err := utils.LoadDescriptorSet("fixtures", "fileset.pb")
			req := utils.CreateGenRequest(pf, protofile)
			fmt.Println(req.FileToGenerate)
			require.NoError(t, err)

			b, err := proto.Marshal(req)
			require.NoError(t, err)

			resp, err := ConvertFrom(bytes.NewBuffer(b))
			require.NoError(t, err)
			assert.Len(t, resp.File, 1)
			assert.NotNil(t, resp.File[0].Name)
			baseFilename := strings.TrimSuffix(protofile, filepath.Ext(protofile))
			assert.Equal(t, baseFilename+".openapi.json", resp.File[0].GetName())
			expectedFile, err := os.ReadFile(baseFilename + ".openapi.json")
			require.NoError(t, err)

			assert.Equal(t, string(expectedFile), resp.File[0].GetContent())
		})
	}
}
