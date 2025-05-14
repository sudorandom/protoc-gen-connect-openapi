package options

import (
	"fmt"
	"os"
	"path"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Options struct {
	// Format is either 'yaml' or 'json' and is the format of the output OpenAPI file(s).
	Format string
	// BaseOpenAPI is the file contents of a base OpenAPI file.
	BaseOpenAPI []byte
	// OverrideOpenAPI is the file contents of an override OpenAPI file.
	OverrideOpenAPI []byte
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
	// PathPrefix is a prefix that is prepended to every HTTP path.
	PathPrefix string
	// TrimUnusedTypes will remove types that aren't referenced by a service.
	TrimUnusedTypes bool
	// WithProtoAnnotations will add some protobuf annotations for descriptions
	WithProtoAnnotations bool
	// FullyQualifiedMessageNames uses the full path for message types: {pkg}.{name} instead of just the name. This
	// is helpful if you are mixing types from multiple services.
	FullyQualifiedMessageNames bool
	// Prevents adding default tags to converted fields
	WithoutDefaultTags bool
	// WithServiceDescriptions set to true will cause service names and their comments to be added to the end of info.description.
	WithServiceDescriptions bool
	// IgnoreGoogleapiHTTP set to true will cause service to always generate OpenAPI specs for connect endpoints, and ignore any google.api.http options.
	IgnoreGoogleapiHTTP bool
	// Services filters which services will be used for generating OpenAPI spec.
	Services []protoreflect.FullName
	// ShortServiceTags uses the short service name (Name()) instead of the full name (FullName()) for OpenAPI tags.
	ShortServiceTags bool
	// ShortOperationIds sets the operationId to shortServiceName + "_" + method short name instead of the full method name.
	ShortOperationIds bool

	MessageAnnotator        MessageAnnotator
	FieldAnnotator          FieldAnnotator
	FieldReferenceAnnotator FieldReferenceAnnotator
}

func (opts Options) HasService(serviceName protoreflect.FullName) bool {
	if len(opts.Services) == 0 {
		return true
	}
	for _, service := range opts.Services {
		if service == serviceName {
			return true
		}
	}
	return false
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
		case param == "fully-qualified-message-names":
			opts.FullyQualifiedMessageNames = true
		case param == "without-default-tags":
			opts.WithoutDefaultTags = true
		case param == "with-service-descriptions":
			opts.WithServiceDescriptions = true
		case param == "ignore-googleapi-http":
			opts.IgnoreGoogleapiHTTP = true
		case param == "short-service-tags":
			opts.ShortServiceTags = true
		case param == "short-operation-ids":
			opts.ShortOperationIds = true
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
		case strings.HasPrefix(param, "path-prefix="):
			opts.PathPrefix = param[12:]
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
		case strings.HasPrefix(param, "override="):
			overridePath := strings.TrimPrefix(param, "override=")
			ext := path.Ext(overridePath)
			switch ext {
			case ".yaml", ".yml", ".json":
				body, err := os.ReadFile(overridePath)
				if err != nil {
					return opts, err
				}
				opts.OverrideOpenAPI = body
			default:
				return opts, fmt.Errorf("the file extension for 'override' should end with yaml or json, not '%s'", ext)
			}
		case strings.HasPrefix(param, "services="):
			services := strings.Split(param[9:], ",")
			for _, service := range services {
				opts.Services = append(opts.Services, protoreflect.FullName(service))
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
