package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestVersionInfo(t *testing.T) {
	v, c, d := getVersionInfo()
	assert.NotEmpty(t, v)
	assert.NotEmpty(t, c)
	assert.NotEmpty(t, d)

	fv := fullVersion()
	assert.NotEmpty(t, fv)
	assert.Contains(t, fv, v)
}

func TestRenderResponse(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	resp := &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{
			{
				Name:    proto.String("test.yaml"),
				Content: proto.String("openapi: 3.1.0"),
			},
		},
	}

	renderResponse(resp)
	require.NoError(t, w.Close())
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	outResp := &pluginpb.CodeGeneratorResponse{}
	err = proto.Unmarshal(buf.Bytes(), outResp)
	require.NoError(t, err)
	require.Len(t, outResp.File, 1)
	assert.Equal(t, "test.yaml", outResp.File[0].GetName())
	assert.Equal(t, "openapi: 3.1.0", outResp.File[0].GetContent())
}
