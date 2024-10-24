package options

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/annotations"
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
	// TrimUnusedTypes will remove types that aren't referenced by a service.
	TrimUnusedTypes bool
	// WithProtoAnnotations will add some protobuf annotations for descriptions
	WithProtoAnnotations bool

	MessageAnnotator        annotations.MessageAnnotator
	FieldAnnotator          annotations.FieldAnnotator
	FieldReferenceAnnotator annotations.FieldReferenceAnnotator
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
		case param == "with-proto-annotations":
			opts.WithProtoAnnotations = true
		case param == "trim-unused-types":
			opts.TrimUnusedTypes = true
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

func IsValidContentType(contentType string) bool {
	for _, protocol := range Protocols {
		if protocol.Name == contentType {
			return true
		}
	}
	return false
}
