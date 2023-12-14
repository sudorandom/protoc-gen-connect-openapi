package main

import (
	"bytes"
	"os"
	"path"
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
	cases := []struct {
		protofile        string
		expectedFilename string
	}{
		{
			protofile:        "helloworld/helloworld.proto",
			expectedFilename: "helloworld/helloworld.openapi.json",
		},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.protofile, func(t *testing.T) {
			pf, err := utils.LoadDescriptorSet("fixtures", "fileset.pb")
			req := utils.CreateGenRequest(pf, tt.protofile)
			require.NoError(t, err)

			b, err := proto.Marshal(req)
			require.NoError(t, err)

			resp, err := ConvertFrom(bytes.NewBuffer(b))
			require.NoError(t, err)
			assert.Len(t, resp.File, 1)
			assert.NotNil(t, resp.File[0].Name)
			assert.Equal(t, path.Join("fixtures", tt.expectedFilename), resp.File[0].GetName())

			baseFilename := strings.TrimSuffix(tt.protofile, filepath.Ext(tt.protofile))
			expectedFile, err := os.ReadFile(path.Join("fixtures", baseFilename+".openapi.json"))
			require.NoError(t, err)

			assert.Equal(t, string(expectedFile), resp.File[0].GetContent())
		})
	}
}
