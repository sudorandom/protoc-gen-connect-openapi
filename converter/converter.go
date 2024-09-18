package converter

import (
	"cmp"
	"fmt"
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
	g, err := generatorWithOptions()
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

// WithBaseOpenAPI sets a file to use as a base for all OpenAPI files.
func WithBaseOpenAPI(baseOpenAPI []byte) Option {
	return func(g *generator) error {
		g.options.BaseOpenAPI = baseOpenAPI
		return nil
	}
}

// WithAllowGET sets a file to use as a base for all OpenAPI files.
func WithAllowGET(allowGet bool) Option {
	return func(g *generator) error {
		g.options.AllowGET = allowGet
		return nil
	}
}

// WithContentTypes sets a file to use as a base for all OpenAPI files.
func WithContentTypes(contentTypes ...string) Option {
	return func(g *generator) error {
		for _, contentType := range contentTypes {
			if !options.IsValidContentType(contentType) {
				return fmt.Errorf("unknown content type: '%s'", contentType)
			}
			g.options.ContentTypes[contentType] = struct{}{}
		}
		return nil
	}
}

// WithIncludeNumberEnumValues sets a file to use as a base for all OpenAPI files.
func WithIncludeNumberEnumValues(includeNumberEnumValues bool) Option {
	return func(g *generator) error {
		g.options.IncludeNumberEnumValues = includeNumberEnumValues
		return nil
	}
}

// WithStreaming sets a file to use as a base for all OpenAPI files.
func WithStreaming(streaming bool) Option {
	return func(g *generator) error {
		g.options.WithStreaming = streaming
		return nil
	}
}
