package main

import (
	"bytes"
	"errors"
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

		formats := []string{}
		for _, format := range []string{"yaml", "json"} {
			_, err := os.Stat(makeOutputPath(protofile, format))
			errors.Is(err, os.ErrNotExist)
			if err != nil && errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				require.NoError(t, err)
			}

			formats = append(formats, format)
		}

		require.NotZero(t, len(formats), "at least one output format has to exist")

		relPath := strings.TrimPrefix(protofile, "fixtures/")
		for _, format := range formats {
			format := format
			t.Run(protofile+"→"+format, func(t *testing.T) {
				pf, err := utils.LoadDescriptorSet("fixtures", "fileset.binpb")
				req := utils.CreateGenRequest(pf, relPath)
				params := fmt.Sprintf("format_%s", format)
				req.Parameter = &params
				require.NoError(t, err)

				b, err := proto.Marshal(req)
				require.NoError(t, err)

				resp, err := ConvertFrom(bytes.NewBuffer(b))
				require.NoError(t, err)
				assert.Len(t, resp.File, 1)
				assert.NotNil(t, resp.File[0].Name)
				assert.Equal(t, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".openapi."+format, resp.File[0].GetName())

				expectedFile, err := os.ReadFile(makeOutputPath(protofile, format))
				require.NoError(t, err)
				assert.Equal(t, string(expectedFile), resp.File[0].GetContent())
			})
		}
	}
}

func makeOutputPath(protofile, format string) string {
	return strings.TrimSuffix(protofile, filepath.Ext(protofile)) + ".openapi." + format
}
