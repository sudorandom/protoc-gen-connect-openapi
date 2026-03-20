package converter

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	intconverter "github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var Convert = intconverter.Convert

type generator struct {
	req     *pluginpb.CodeGeneratorRequest
	options options.Options
}

// Generate a single OpenAPI file.
func GenerateSingle(opts ...Option) ([]byte, error) {
	g, err := generatorWithOptions(opts...)
	if err != nil {
		return nil, err
	}
	g.options.Path = "all"
	resp, err := intconverter.ConvertWithOptions(g.req, g.options)
	if err != nil {
		return nil, err
	}
	return []byte(resp.File[0].GetContent()), nil
}

// Generate OpenAPI files with the given options.
func Generate(opts ...Option) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	g, err := generatorWithOptions(opts...)
	if err != nil {
		return nil, err
	}
	resp, err := intconverter.ConvertWithOptions(g.req, g.options)
	if err != nil {
		return nil, err
	}
	return resp.GetFile(), nil
}

func generatorWithOptions(opts ...Option) (*generator, error) {
	g := &generator{
		req: &pluginpb.CodeGeneratorRequest{
			FileToGenerate:        []string{},
			Parameter:             new(string),
			ProtoFile:             []*descriptorpb.FileDescriptorProto{},
			SourceFileDescriptors: []*descriptorpb.FileDescriptorProto{},
			CompilerVersion:       &pluginpb.Version{},
		},
		options: options.NewOptions(),
	}
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}
	return g, nil
}

type Option func(*generator) error

// WithSourceFiles adds the given files as source files but won't generate OpenAPI based on any services found in here.
func WithSourceFiles(files *protoregistry.Files) Option {
	return func(g *generator) error {
		return withSourceFiles(files, g)
	}
}

// WithFiles will generate OpenAPI specs for the given files.
func WithFiles(files *protoregistry.Files) Option {
	return func(g *generator) error {
		files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			if fd.Services().Len() > 0 {
				g.req.FileToGenerate = append(g.req.FileToGenerate, string(fd.Path()))
			}
			return true
		})
		slices.Sort(g.req.FileToGenerate)
		if err := withSourceFiles(files, g); err != nil {
			return err
		}
		return nil
	}
}

func withSourceFiles(files *protoregistry.Files, g *generator) error {
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		g.req.ProtoFile = append(g.req.ProtoFile, protodesc.ToFileDescriptorProto(fd.ParentFile()))
		return true
	})
	slices.SortFunc(g.req.ProtoFile, func(a *descriptorpb.FileDescriptorProto, b *descriptorpb.FileDescriptorProto) int {
		return cmp.Compare(a.GetPackage(), b.GetPackage())
	})
	return nil
}

// WithGlobal will generate OpenAPI specs for any service in the global registry. Shortcut for converter.WithFiles(protoregistry.GlobalFiles).
func WithGlobal() Option {
	return WithFiles(protoregistry.GlobalFiles)
}

// WithFormat sets the format for the OpenAPI file.
func WithFormat(format string) Option {
	return func(g *generator) error {
		g.options.Format = format
		return nil
	}
}

// WithBaseOpenAPI sets a base OpenAPI document to merge into the generated output.
func WithBaseOpenAPI(baseOpenAPI []byte) Option {
	return func(g *generator) error {
		g.options.BaseOpenAPI = baseOpenAPI
		return nil
	}
}

// WithOverrideOpenAPI sets an override OpenAPI document to merge on top of the generated output.
func WithOverrideOpenAPI(overrideOpenAPI []byte) Option {
	return func(g *generator) error {
		g.options.OverrideOpenAPI = overrideOpenAPI
		return nil
	}
}

// WithAllowGET allows methods with idempotency_level = NO_SIDE_EFFECTS to be documented with GET requests.
func WithAllowGET(allowGet bool) Option {
	return func(g *generator) error {
		g.options.AllowGET = allowGet
		return nil
	}
}

// WithContentTypes sets the content types to include in the generated OpenAPI output.
func WithContentTypes(contentTypes ...string) Option {
	return func(g *generator) error {
		g.options.ContentTypes = map[string]struct{}{}
		for _, contentType := range contentTypes {
			if !options.IsValidContentType(contentType) {
				return fmt.Errorf("unknown content type: '%s'", contentType)
			}
			g.options.ContentTypes[contentType] = struct{}{}
		}
		return nil
	}
}

// WithIncludeNumberEnumValues includes numeric values for enums in addition to string representations.
func WithIncludeNumberEnumValues(includeNumberEnumValues bool) Option {
	return func(g *generator) error {
		g.options.IncludeNumberEnumValues = includeNumberEnumValues
		return nil
	}
}

// WithIgnoreGoogleapiHTTP tells the generator to ignore google.api.http options.
func WithIgnoreGoogleapiHTTP(ignoreGoogleapiHTTP bool) Option {
	return func(g *generator) error {
		g.options.IgnoreGoogleapiHTTP = ignoreGoogleapiHTTP
		return nil
	}
}

// WithOnlyGoogleapiHTTP tells the generator to only include methods with google.api.http options.
func WithOnlyGoogleapiHTTP(onlyGoogleapiHTTP bool) Option {
	return func(g *generator) error {
		g.options.OnlyGoogleapiHTTP = onlyGoogleapiHTTP
		return nil
	}
}

// WithStreaming includes content types related to streaming.
func WithStreaming(streaming bool) Option {
	return func(g *generator) error {
		g.options.WithStreaming = streaming
		return nil
	}
}

// WithDebug sets up the logger to emit debug entries
func WithDebug(enabled bool) Option {
	return func(g *generator) error {
		g.options.Debug = enabled
		return nil
	}
}

// WithProtoNames uses protobuf field names instead of JSON names.
func WithProtoNames(enabled bool) Option {
	return func(g *generator) error {
		g.options.WithProtoNames = enabled
		return nil
	}
}

// WithTrimUnusedTypes removes types that aren't referenced by a service.
func WithTrimUnusedTypes(enabled bool) Option {
	return func(g *generator) error {
		g.options.TrimUnusedTypes = enabled
		return nil
	}
}

// WithProtoAnnotations adds some details about protobuf to descrioptions.
func WithProtoAnnotations(enabled bool) Option {
	return func(g *generator) error {
		g.options.WithProtoAnnotations = enabled
		return nil
	}
}

// WithServices will limit the services generated.
func WithServices(serviceNames []protoreflect.FullName) Option {
	return func(g *generator) error {
		serviceNameStrs := make([]string, len(serviceNames))
		for i, serviceName := range serviceNames {
			serviceNameStrs[i] = string(serviceName)
		}
		services, err := options.CompileServicePatterns(serviceNameStrs)
		if err != nil {
			return fmt.Errorf("invalid service patterns: %w", err)
		}
		g.options.Services = append(g.options.Services, services...)
		return nil
	}
}

// WithServicePatterns will limit the services generated using glob patterns "company.service.*Service"
func WithServicePatterns(serviceNames []string) Option {
	return func(g *generator) error {
		services, err := options.CompileServicePatterns(serviceNames)
		if err != nil {
			return fmt.Errorf("invalid service patterns: %w", err)
		}
		g.options.Services = append(g.options.Services, services...)

		return nil
	}
}

// WithShortServiceTags uses the short service name instead of the full name for OpenAPI tags.
func WithShortServiceTags(enabled bool) Option {
	return func(g *generator) error {
		g.options.ShortServiceTags = enabled
		return nil
	}
}

// WithShortOperationIds sets the operationId to shortServiceName + "_" + method short name instead of the full method name.
func WithShortOperationIds(enabled bool) Option {
	return func(g *generator) error {
		g.options.ShortOperationIds = enabled
		return nil
	}
}

// WithoutDefaultTags prevents adding default tags to converted fields.
func WithoutDefaultTags(enabled bool) Option {
	return func(g *generator) error {
		g.options.WithoutDefaultTags = enabled
		return nil
	}
}

// WithDisableDefaultResponse disables the default 200 response.
func WithDisableDefaultResponse(enabled bool) Option {
	return func(g *generator) error {
		g.options.DisableDefaultResponse = enabled
		return nil
	}
}

// WithAllowedVisibilities sets the visibility restriction labels to include.
func WithAllowedVisibilities(visibilities ...string) Option {
	return func(g *generator) error {
		g.options.AllowedVisibilities = make(map[string]bool)
		for _, v := range visibilities {
			g.options.AllowedVisibilities[v] = true
		}
		return nil
	}
}

// WithFullyQualifiedMessageNames decides if you want to use the full path in message names.
func WithFullyQualifiedMessageNames(enabled bool) Option {
	return func(g *generator) error {
		g.options.FullyQualifiedMessageNames = enabled
		return nil
	}
}

// WithServiceDescriptions decides if service names and their comments to be added to the end of info.description.
func WithServiceDescriptions(enabled bool) Option {
	return func(g *generator) error {
		g.options.WithServiceDescriptions = enabled
		return nil
	}
}

// WithPathPrefix prepends a given string to each HTTP path.
func WithPathPrefix(prefix string) Option {
	return func(g *generator) error {
		g.options.PathPrefix = prefix
		return nil
	}
}

// WithGoogleErrorDetail enables the generation of error details using error_details.proto from google.rpc.
func WithGoogleErrorDetail(enabled bool) Option {
	return func(g *generator) error {
		g.options.WithGoogleErrorDetail = enabled
		return nil
	}
}

// WithLogger sets the logger to a given *slog.Logger instance. The default behavior will discard logs.
func WithLogger(logger *slog.Logger) Option {
	return func(g *generator) error {
		g.options.Logger = logger
		return nil
	}
}

// WithFeatures sets the features that are enabled.
func WithFeatures(features ...options.Feature) Option {
	return func(g *generator) error {
		return g.options.EnableFeatures(features...)
	}
}
