package options

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/gobwas/glob"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Feature string

const (
	FeatureGoogleAPIHTTP Feature = "google.api.http"
	FeatureConnectRPC    Feature = "connectrpc"
	FeatureTwirp         Feature = "twirp"
	FeatureGnostic       Feature = "gnostic"
	FeatureProtovalidate Feature = "protovalidate"
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
	// OnlyGoogleapiHTTP set to true will only generate OpenAPI specs for methods with explicit google.api.http annotations. Methods without google.api.http will be skipped.
	OnlyGoogleapiHTTP bool
	// Services filters which services will be used for generating OpenAPI spec.
	Services []glob.Glob
	// ShortServiceTags uses the short service name (Name()) instead of the full name (FullName()) for OpenAPI tags.
	ShortServiceTags bool
	// ShortOperationIds sets the operationId to shortServiceName + "_" + method short name instead of the full method name.
	ShortOperationIds bool
	// WithGoogleErrorDetail will add google error detail to the connect error response.
	WithGoogleErrorDetail bool
	// DisableDefaultResponse disables the default 200 response.
	DisableDefaultResponse bool
	// EnabledFeatures is a map of enabled features.
	EnabledFeatures map[Feature]bool
	// AllowedVisibilities is a map of visibility strings to include. If an element has a `google.api.visibility` rule with a `restriction` that is not in this map, it will be excluded.
	AllowedVisibilities map[string]bool

	MessageAnnotator        MessageAnnotator
	FieldAnnotator          FieldAnnotator
	FieldReferenceAnnotator FieldReferenceAnnotator

	ExtensionTypeResolver protoregistry.ExtensionTypeResolver

	Logger *slog.Logger
}

func (opts Options) FeatureEnabled(feature Feature) bool {
	return opts.EnabledFeatures[feature]
}

func (opts Options) HasService(serviceName protoreflect.FullName) bool {
	if len(opts.Services) == 0 {
		return true
	}
	for _, pattern := range opts.Services {
		if pattern.Match(string(serviceName)) {
			return true
		}
	}
	return false
}

func (opts *Options) EnableFeatures(features ...Feature) error {
	enabledFeatures := make(map[Feature]bool)
	for _, feature := range features {
		switch feature {
		case FeatureGoogleAPIHTTP, FeatureConnectRPC, FeatureTwirp, FeatureGnostic, FeatureProtovalidate:
			enabledFeatures[feature] = true
		default:
			return fmt.Errorf("invalid feature: '%s'", feature)
		}
	}
	opts.EnabledFeatures = enabledFeatures
	return nil
}

func NewOptions() Options {
	return Options{
		Format: "yaml",
		ContentTypes: map[string]struct{}{
			"json": {},
		},
		EnabledFeatures: map[Feature]bool{
			FeatureConnectRPC:    true,
			FeatureGoogleAPIHTTP: true,
			FeatureGnostic:       true,
			FeatureProtovalidate: true,
		},
		Logger: slog.New(slog.DiscardHandler), // discard logs by default,
	}
}

func (opts Options) GetExtensionTypeResolver() protoregistry.ExtensionTypeResolver {
	if opts.ExtensionTypeResolver == nil {
		return protoregistry.GlobalTypes
	}
	return opts.ExtensionTypeResolver
}

func FromString(s string) (Options, error) {
	opts := NewOptions()

	supportedProtocols := map[string]struct{}{}
	for _, proto := range Protocols {
		supportedProtocols[proto.Name] = struct{}{}
	}

	contentTypes := map[string]struct{}{}
	for param := range strings.SplitSeq(s, ",") {
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
		case param == "only-googleapi-http":
			opts.OnlyGoogleapiHTTP = true
		case param == "short-service-tags":
			opts.ShortServiceTags = true
		case param == "short-operation-ids":
			opts.ShortOperationIds = true
		case param == "with-google-error-detail":
			opts.WithGoogleErrorDetail = true
		case param == "disable-default-response":
			opts.DisableDefaultResponse = true
		case strings.HasPrefix(param, "features="):
			allFeatures := []Feature{}
			for feature := range strings.SplitSeq(param[9:], ";") {
				feature = strings.TrimSpace(feature)
				allFeatures = append(allFeatures, Feature(feature))
			}

			err := opts.EnableFeatures(allFeatures...)
			if err != nil {
				return opts, err
			}
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
			if msg, ok := disabledOptions["base"]; ok {
				return opts, errors.New(msg)
			}
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
			if msg, ok := disabledOptions["override"]; ok {
				return opts, errors.New(msg)
			}
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
			patterns, err := CompileServicePatterns(services)
			if err != nil {
				return opts, err
			}
			opts.Services = append(opts.Services, patterns...)
		case strings.HasPrefix(param, "allowed-visibilities="):
			selectors := strings.Split(param[len("allowed-visibilities="):], ";")
			opts.AllowedVisibilities = make(map[string]bool)
			for _, selector := range selectors {
				opts.AllowedVisibilities[selector] = true
			}
		default:
			return opts, fmt.Errorf("invalid parameter: %s", param)
		}
	}
	if len(contentTypes) > 0 {
		opts.ContentTypes = contentTypes
	}
	if opts.IgnoreGoogleapiHTTP {
		opts.Logger.Debug("Ignoring google.api.http")
		opts.EnabledFeatures[FeatureGoogleAPIHTTP] = false
	}
	if opts.OnlyGoogleapiHTTP {
		opts.Logger.Debug("Only google.api.http enabled")
		opts.EnabledFeatures[FeatureConnectRPC] = false
		opts.EnabledFeatures[FeatureTwirp] = false
		opts.EnabledFeatures[FeatureGoogleAPIHTTP] = true
	}
	opts.Logger.Debug("Enabled features before final check", "features", opts.EnabledFeatures)
	hasProtocolFeature := opts.FeatureEnabled(FeatureConnectRPC) || opts.FeatureEnabled(FeatureGoogleAPIHTTP) || opts.FeatureEnabled(FeatureTwirp)
	if !hasProtocolFeature {
		return opts, errors.New("at least one protocol feature (connectrpc, google.api.http, or twirp) must be enabled")
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

func CompileServicePatterns(services []string) ([]glob.Glob, error) {
	var patterns []glob.Glob
	for _, service := range services {
		pattern, err := glob.Compile(service, '.')
		if err != nil {
			return patterns, fmt.Errorf("invalid service glob pattern '%s': %w", service, err)
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}
