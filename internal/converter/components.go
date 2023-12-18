package converter

import (
	"log/slog"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Protocol struct {
	ContentType  string
	RequestDesc  string
	ResponseDesc string
	IsStreaming  bool
	IsBinary     bool
}

var Protocols = []Protocol{
	{
		// No neeed to explain JSON :)
		ContentType: "application/json",
	},
	{
		ContentType:  "application/proto",
		RequestDesc:  "The request is binary-concoded protobuf without gRPC framing. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is binary-concoded protobuf without gRPC framing. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsBinary:     true,
	},
	{
		ContentType:  "application/connect+json",
		RequestDesc:  "The request is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/connect+proto",
		RequestDesc:  "The request is binary-concoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is binary-concoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc+proto",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc+json",
		RequestDesc:  "The request is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc-web",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc-web+proto",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		ContentType:  "application/grpc-web+json",
		RequestDesc:  "The request is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
}

func fileToComponents(fd protoreflect.FileDescriptor) (openapi31.Components, error) {
	// Add schema from messages/enums
	components := openapi31.Components{}
	st := NewState()
	slog.Info("start collection")
	st.CollectFile(fd)
	slog.Info("collection complete", slog.String("file", string(fd.Name())), slog.Int("messages", len(st.Messages)), slog.Int("enum", len(st.Enums)))
	rootSchema := stateToSchema(st)
	for _, item := range rootSchema.Items.SchemaArray {
		if item.TypeObject == nil {
			continue
		}
		m, err := item.ToSimpleMap()
		if err != nil {
			return components, err
		}
		// We don't actually want to use the $id property so clear it out and just use it in the path
		delete(m, "$id")
		components.WithSchemasItem(*item.TypeObject.ID, m)
	}

	// Add requestBodies and responses for methods
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
			op := &openapi31.Operation{}
			op.WithTags(string(service.FullName()))
			loc := fd.SourceLocations().ByDescriptor(method)
			op.WithDescription(formatComments(loc))

			// Request Body
			if !IsEmpty(method.Input()) {
				inputRef := "#/components/schemas/" + formatTypeRef(string(method.Input().FullName()))
				components.WithRequestBodiesItem(formatTypeRef(string(method.Input().FullName())),
					openapi31.RequestBodyOrReference{
						RequestBody: &openapi31.RequestBody{
							Content:  makeMediaTypes(inputRef, true, isStreaming),
							Required: BoolPtr(true),
						},
					},
				)
			}

			if !IsEmpty(method.Output()) {
				outputRef := "#/components/schemas/" + formatTypeRef(string(method.Output().FullName()))
				components.WithResponsesItem(formatTypeRef(string(method.Output().FullName())),
					openapi31.ResponseOrReference{
						Response: &openapi31.Response{
							Content: makeMediaTypes(outputRef, false, isStreaming),
						},
					},
				)
			}
		}
	}

	// Add our own type for errors
	reflector := jsonschema.Reflector{}
	connectError, err := reflector.Reflect(ConnectError{})
	if err != nil {
		return components, err
	}
	connectError.WithTitle("Connect Error")
	connectError.WithDescription(`Error type returned by Connect: https://connectrpc.com/docs/go/errors/#http-representation`)
	connectError.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: BoolPtr(false)})

	components.WithSchemasItem("connect.error", map[string]interface{}{
		"description":          connectError.Description,
		"properties":           connectError.Properties,
		"title":                connectError.Title,
		"type":                 connectError.Type,
		"additionalProperties": connectError.AdditionalProperties,
	})

	components.WithResponsesItem("connect.error", openapi31.ResponseOrReference{
		Response: &openapi31.Response{
			Content: makeMediaTypes("#/components/schemas/connect.error", false, false),
		},
	})

	return components, nil
}

// makeMediaTypes generates media types with references to the bodies
func makeMediaTypes(ref string, isRequest, isStreaming bool) map[string]openapi31.MediaType {
	mediaTypes := map[string]openapi31.MediaType{}
	for _, protocol := range Protocols {
		// Make sure the protocol supports streaming if this is a streaming RPC
		if isStreaming && !protocol.IsStreaming {
			continue
		}

		var description string
		if isRequest {
			description = protocol.RequestDesc
		} else {
			description = protocol.ResponseDesc
		}

		// Since this protocol has a description, wrap it
		if description != "" {
			mediaTypes[protocol.ContentType] = openapi31.MediaType{
				Schema: map[string]interface{}{
					"type":   jsonschema.String.Type(),
					"format": "binary",
					"properties": map[string]jsonschema.SchemaOrBool{
						"contents": {
							TypeObject: (&jsonschema.Schema{}).WithRef(ref),
						},
					},
					"description": description,
				},
			}
		} else {
			mediaTypes[protocol.ContentType] = openapi31.MediaType{
				Schema: map[string]interface{}{
					"$ref": ref,
				},
			}
		}
	}
	return mediaTypes
}

// ConnectError is an error that
type ConnectError struct {
	Code    string `json:"code" example:"CodeNotFound" enum:"CodeCanceled,CodeUnknown,CodeInvalidArgument,CodeDeadlineExceeded,CodeNotFound,CodeAlreadyExists,CodePermissionDenied,CodeResourceExhausted,CodeFailedPrecondition,CodeAborted,CodeOutOfRange,CodeInternal,CodeUnavailable,CodeDataLoss,CodeUnauthenticated"`
	Message string `json:"message,omitempty"`
}
