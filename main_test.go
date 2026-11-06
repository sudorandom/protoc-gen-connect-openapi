// Copyright 2023-2026 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"errors"
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
	resp := &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{
			{
				Name:    proto.String("test.yaml"),
				Content: proto.String("openapi: 3.1.0"),
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, renderResponse(&buf, resp))

	outResp := &pluginpb.CodeGeneratorResponse{}
	require.NoError(t, proto.Unmarshal(buf.Bytes(), outResp))
	require.Len(t, outResp.File, 1)
	assert.Equal(t, "test.yaml", outResp.File[0].GetName())
	assert.Equal(t, "openapi: 3.1.0", outResp.File[0].GetContent())
}

func TestRunReportsConversionErrorsInResponse(t *testing.T) {
	req, err := proto.Marshal(&pluginpb.CodeGeneratorRequest{
		Parameter: proto.String("not-a-real-option"),
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, run(bytes.NewReader(req), &buf))

	outResp := &pluginpb.CodeGeneratorResponse{}
	require.NoError(t, proto.Unmarshal(buf.Bytes(), outResp))
	assert.Contains(t, outResp.GetError(), "not-a-real-option")
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

func TestRunReturnsWriteErrors(t *testing.T) {
	req, err := proto.Marshal(&pluginpb.CodeGeneratorRequest{})
	require.NoError(t, err)
	assert.Error(t, run(bytes.NewReader(req), failingWriter{}))
}
