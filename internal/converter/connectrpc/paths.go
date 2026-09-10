package connectrpc

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
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

	operationId := string(method.FullName())
	if opts.ShortOperationIds {
		operationId = string(service.Name()) + "_" + string(method.Name())
	}

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

	op.Responses.Default = &v3.Response{
		Description: "Error",
		Content: util.MakeMediaTypes(
			opts,
			base.CreateSchemaProxyRef("#/components/schemas/connect.error"),
			false,
			isStreaming,
		),
	}
	op.Parameters = append(op.Parameters,
		&v3.Parameter{
			Name:        "Connect-Protocol-Version",
			In:          "header",
			Description: "Define the version of the Connect protocol",
			Required:    util.BoolPtr(true),
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
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/"+util.FormatTypeRef(inputId)),
					true,
					isStreaming),
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
