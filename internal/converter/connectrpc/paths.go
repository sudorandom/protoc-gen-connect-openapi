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

package connectrpc

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func MakePathItems(opts options.Options, service protoreflect.ServiceDescriptor, method protoreflect.MethodDescriptor) *orderedmap.Map[string, *v3.PathItem] {
	path := "/" + string(service.FullName()) + "/" + string(method.Name())
	paths := orderedmap.New[string, *v3.PathItem]()
	item := &v3.PathItem{}
	hasGetSupport := methodHasGet(opts, method)
	if hasGetSupport {
		item.Get = MethodToOperation(opts, method, true)
	}
	item.Post = MethodToOperation(opts, method, false)
	// Streaming methods produce no operations without with-streaming; an empty
	// path item would still show up as a route.
	if item.Get == nil && item.Post == nil {
		return paths
	}
	paths.Set(path, item)
	return paths
}

func methodHasGet(opts options.Options, method protoreflect.MethodDescriptor) bool {
	if !opts.AllowGET {
		return false
	}

	if method.IsStreamingClient() || method.IsStreamingServer() {
		return false
	}

	options := method.Options().(*descriptorpb.MethodOptions)
	return options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS
}

func MethodToOperation(opts options.Options, method protoreflect.MethodDescriptor, returnGet bool) *v3.Operation {
	fd := method.ParentFile()
	service := method.Parent().(protoreflect.ServiceDescriptor)
	loc := fd.SourceLocations().ByDescriptor(method)
	tagName := string(service.FullName())
	if opts.ShortServiceTags {
		tagName = string(service.Name())
	}

	operationId := util.DefaultOperationID(opts, method)

	summary, description := util.FormatOperationComments(loc)
	if summary == "" {
		summary = string(method.Name())
	}
	op := &v3.Operation{
		Summary:     summary,
		OperationId: operationId,
		Deprecated:  util.IsMethodDeprecated(method),
		Tags:        []string{tagName},
		Description: description,
	}

	isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
	if isStreaming && !opts.WithStreaming {
		return nil
	}

	// Responses
	op.Responses = &v3.Responses{
		Codes: orderedmap.New[string, *v3.Response](),
	}
	if !opts.DisableDefaultResponse {
		outputId := util.FormatTypeRef(string(method.Output().FullName()))
		op.Responses.Codes.Set("200", &v3.Response{
			Description: "Success",
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/"+outputId),
				false,
				isStreaming,
			),
		})
	}

	// Unary Connect errors are always JSON, whatever codec the request used.
	// Streaming errors arrive in the stream's end message, in the stream's
	// content type.
	errorContent := util.MakeMediaTypes(
		opts,
		base.CreateSchemaProxyRef("#/components/schemas/connect.error"),
		false,
		isStreaming,
	)
	if !isStreaming {
		errorContent = jsonOnly(base.CreateSchemaProxyRef("#/components/schemas/connect.error"))
	}
	op.Responses.Default = &v3.Response{
		Description: "Error",
		Content:     errorContent,
	}
	op.Parameters = append(op.Parameters,
		&v3.Parameter{
			Name:        "Connect-Protocol-Version",
			In:          "header",
			Description: "Define the version of the Connect protocol",
			Schema:      base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
			Example:     utils.CreateIntNode("1"),
		},
		&v3.Parameter{
			Name:        "Connect-Timeout-Ms",
			In:          "header",
			Description: "Define the timeout, in ms",
			Schema:      base.CreateSchemaProxyRef("#/components/schemas/connect-timeout-header"),
			Example:     utils.CreateIntNode("1000"),
		},
	)

	// Request parameters
	inputId := util.FormatTypeRef(string(method.Input().FullName()))
	if returnGet {
		op.OperationId = op.OperationId + ".get"
		op.Parameters = append(op.Parameters,
			&v3.Parameter{
				Name:        "message",
				In:          "query",
				Description: "The URL-encoded message query parameter contains the JSON or binary serialized request message",
				// A parameter's content map must have exactly one entry, so the
				// message is described in its JSON form (encoding=json).
				Content: jsonOnly(base.CreateSchemaProxyRef("#/components/schemas/" + util.FormatTypeRef(inputId))),
			},
			&v3.Parameter{
				Name:        "encoding",
				In:          "query",
				Description: "Define which encoding or 'Message-Codec' to use",
				Required:    util.BoolPtr(true),
				Schema:      base.CreateSchemaProxyRef("#/components/schemas/encoding"),
				Example:     utils.CreateStringNode("json"),
			},
			&v3.Parameter{
				Name:        "base64",
				In:          "query",
				Description: "Specifies if the message query param is base64 encoded, which may be required for binary data",
				Schema:      base.CreateSchemaProxyRef("#/components/schemas/base64"),
				Example:     utils.CreateBoolNode("false"),
			},
			&v3.Parameter{
				Name:        "compression",
				In:          "query",
				Description: "Which compression algorithm to use for this request",
				Schema:      base.CreateSchemaProxyRef("#/components/schemas/compression"),
				Example:     utils.CreateStringNode("gzip"),
			},
			&v3.Parameter{
				Name:        "connect",
				In:          "query",
				Description: "Define the version of the Connect protocol",
				Required:    util.BoolPtr(true),
				Schema:      base.CreateSchemaProxyRef("#/components/schemas/connect"),
				Example:     utils.CreateStringNode("v1"),
			},
		)
	} else {
		op.RequestBody = &v3.RequestBody{
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/"+inputId),
				true,
				isStreaming,
			),
			Required: util.BoolPtr(true),
		}
	}

	return op
}

func jsonOnly(schema *base.SchemaProxy) *orderedmap.Map[string, *v3.MediaType] {
	content := orderedmap.New[string, *v3.MediaType]()
	content.Set("application/json", &v3.MediaType{Schema: schema})
	return content
}
