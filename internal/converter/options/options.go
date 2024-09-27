package options

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type Options struct {
	// Format is either 'yaml' or 'json' and is the format of the output OpenAPI file(s).
	Format string
	// BaseOpenAPI is the file contents of a base OpenAPI file.
	BaseOpenAPI []byte
	// WithStreaming will content types related to streaming (warning: can be messy).
	WithStreaming bool
	// AllowGET will let methods with `idempotency_level = NO_SIDE_EFFECTS` to be documented with GET requests.
	AllowGET bool
	// ContentTypes is a map of all content types. Available values are in Protocols.
	ContentTypes map[string]struct{}
	// Debug enables debug logging if set to true.
	Debug bool
	// IncludeNumberEnumValues indicates if numbers are included for enum values in addition to the string representations.
	IncludeNumberEnumValues bool
	// WithProtoNames indicates if protobuf field names should be used instead of JSON names.
	WithProtoNames bool
	// Path is the output OpenAPI path.
	Path string
}

func NewOptions() Options {
	return Options{
		Format: "yaml",
		ContentTypes: map[string]struct{}{
			"json": {},
		},
	}
}

func FromString(s string) (Options, error) {
	opts := NewOptions()

	supportedProtocols := map[string]struct{}{}
	for _, proto := range Protocols {
		supportedProtocols[proto.Name] = struct{}{}
	}

	contentTypes := map[string]struct{}{}
	for _, param := range strings.Split(s, ",") {
		switch {
		case param == "":
		case param == "debug":
			opts.Debug = true
		case param == "include-number-enum-values":
			opts.IncludeNumberEnumValues = true
		case param == "allow-get":
			opts.AllowGET = true
		case param == "with-streaming":
			opts.WithStreaming = true
		case param == "with-proto-names":
			opts.WithProtoNames = true
		case strings.HasPrefix(param, "content-types="):
			for _, contentType := range strings.Split(param[14:], ";") {
				contentType = strings.TrimSpace(contentType)
				_, isSupportedProtocol := supportedProtocols[contentType]
				if !isSupportedProtocol {
					return opts, fmt.Errorf("invalid content type: '%s'", contentType)
				}
				contentTypes[contentType] = struct{}{}
			}
		case strings.HasPrefix(param, "path="):
			opts.Path = param[5:]
		case strings.HasPrefix(param, "format="):
			format := param[7:]
			switch format {
			case "yaml":
				opts.Format = "yaml"
			case "json":
				opts.Format = "json"
			default:
				return opts, fmt.Errorf("format be yaml or json, not '%s'", format)
			}
		case strings.HasPrefix(param, "base="):
			basePath := param[5:]
			ext := path.Ext(basePath)
			switch ext {
			case ".yaml", ".yml", ".json":
				body, err := os.ReadFile(basePath)
				if err != nil {
					return opts, err
				}
				opts.BaseOpenAPI = body
			default:
				return opts, fmt.Errorf("the file extension for 'base' should end with yaml or json, not '%s'", ext)
			}
		default:
			return opts, fmt.Errorf("invalid parameter: %s", param)
		}
	}
	if len(contentTypes) > 0 {
		opts.ContentTypes = contentTypes
	}
	return opts, nil
}

type Protocol struct {
	Name         string
	ContentType  string
	RequestDesc  string
	ResponseDesc string
	IsStreaming  bool
	IsBinary     bool
}

var Protocols = []Protocol{
	{
		// No need to explain JSON :)
		Name:        "json",
		ContentType: "application/json",
	},
	{
		Name:        "proto",
		ContentType: "application/proto",
		IsBinary:    true,
	},
	{
		Name:         "connect+json",
		ContentType:  "application/connect+json",
		RequestDesc:  "The request is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "connect+proto",
		ContentType:  "application/connect+proto",
		RequestDesc:  "The request is binary-encoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is binary-encoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc",
		ContentType:  "application/grpc",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc+proto",
		ContentType:  "application/grpc+proto",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc+json",
		ContentType:  "application/grpc+json",
		RequestDesc:  "The request is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web",
		ContentType:  "application/grpc-web",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web+proto",
		ContentType:  "application/grpc-web+proto",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web+json",
		ContentType:  "application/grpc-web+json",
		RequestDesc:  "The request is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
}

func IsValidContentType(contentType string) bool {
	for _, protocol := range Protocols {
		if protocol.Name == contentType {
			return true
		}
	}
	return false
}
