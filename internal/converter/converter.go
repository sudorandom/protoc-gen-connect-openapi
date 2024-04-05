package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

type Options struct {
	// Format can be either "yaml" or "json"
	Path                    string
	Format                  string
	BaseOpenAPIYAMLPath     string
	BaseOpenAPIJSONPath     string
	WithStreaming           bool
	AllowGET                bool
	ContentTypes            map[string]struct{}
	Debug                   bool
	IncludeNumberEnumValues bool
}

func parseOptions(s string) (Options, error) {
	opts := Options{
		Format:       "yaml",
		ContentTypes: map[string]struct{}{},
	}

	supportedProtocols := map[string]struct{}{}
	for _, proto := range Protocols {
		supportedProtocols[proto.Name] = struct{}{}
	}

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
		case strings.HasPrefix(param, "content-types="):
			for _, contentType := range strings.Split(param[14:], ";") {
				contentType = strings.TrimSpace(contentType)
				_, isSupportedProtocol := supportedProtocols[contentType]
				if !isSupportedProtocol {
					return opts, fmt.Errorf("invalid content type: '%s'", contentType)
				}
				opts.ContentTypes[contentType] = struct{}{}
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
			case ".yaml", ".yml":
				opts.BaseOpenAPIYAMLPath = basePath
			case ".json":
				opts.BaseOpenAPIJSONPath = basePath
			default:
				return opts, fmt.Errorf("the file extension for 'base' should end with yaml or json, not '%s'", ext)
			}
		default:
			return opts, fmt.Errorf("invalid parameter: %s", param)
		}
	}
	if len(opts.ContentTypes) == 0 {
		opts.ContentTypes = map[string]struct{}{
			"json":  {},
			"proto": {},
		}
	}
	return opts, nil
}

func ConvertFrom(rd io.Reader) (*pluginpb.CodeGeneratorResponse, error) {
	input, err := io.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(input, req)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal input: %w", err)
	}

	return Convert(req)
}

func Convert(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	opts, err := parseOptions(req.GetParameter())
	if err != nil {
		return nil, err
	}

	logLevel := slog.LevelInfo
	if opts.Debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: logLevel,
		}),
	))

	files := []*pluginpb.CodeGeneratorResponse_File{}
	genFiles := make(map[string]struct{}, len(req.FileToGenerate))
	for _, file := range req.FileToGenerate {
		genFiles[file] = struct{}{}
	}

	// We need this to resolve dependencies when making protodesc versions of the files
	resolver, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: req.GetProtoFile(),
	})
	if err != nil {
		return nil, err
	}

	newSpec := func() openapi31.Spec {
		return openapi31.Spec{
			Openapi: "3.1.0",
			Info:    openapi31.Info{},
			Paths: &openapi31.Paths{
				MapOfPathItemValues: map[string]openapi31.PathItem{},
				MapOfAnything:       map[string]interface{}{},
			},
			Components: &openapi31.Components{
				Schemas:         map[string]map[string]interface{}{},
				Responses:       map[string]openapi31.ResponseOrReference{},
				Parameters:      map[string]openapi31.ParameterOrReference{},
				Examples:        map[string]openapi31.ExampleOrReference{},
				RequestBodies:   map[string]openapi31.RequestBodyOrReference{},
				Headers:         map[string]openapi31.HeaderOrReference{},
				SecuritySchemes: map[string]openapi31.SecuritySchemeOrReference{},
				Links:           map[string]openapi31.LinkOrReference{},
				Callbacks:       map[string]openapi31.CallbacksOrReference{},
				PathItems:       map[string]openapi31.PathItemOrReference{},
			},
		}
	}

	spec := newSpec()
	outFiles := map[string]openapi31.Spec{}

	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			slog.Debug("skip generating file because it wasn't requested", slog.String("name", fileDesc.GetName()))
			continue
		}

		slog.Debug("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := protodesc.NewFile(fileDesc, resolver)
		if err != nil {
			slog.Error("error loading file", slog.Any("error", err))
			return nil, err
		}

		// Create a per-file openapi spec if we're not merging all into one
		if opts.Path == "" {
			spec = newSpec()
			spec.SetTitle(string(fd.FullName()))
			spec.SetDescription(util.FormatComments(fd.SourceLocations().ByDescriptor(fd)))
		}

		if err := appendToSpec(opts, &spec, fd); err != nil {
			return nil, err
		}

		if opts.Path == "" {
			name := fileDesc.GetName()
			filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".openapi." + opts.Format
			outFiles[filename] = spec
		}
	}

	if opts.Path != "" {
		outFiles[opts.Path] = spec
	}

	for path, spec := range outFiles {
		path := path
		spec := spec
		content, err := specToFile(opts, spec)
		if err != nil {
			return nil, err
		}
		files = append(files, &pluginpb.CodeGeneratorResponse_File{
			Name:              &path,
			Content:           &content,
			GeneratedCodeInfo: &descriptorpb.GeneratedCodeInfo{},
		})
	}

	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	return &pluginpb.CodeGeneratorResponse{
		File:              files,
		SupportedFeatures: &features,
	}, nil
}

func specToFile(opts Options, spec openapi31.Spec) (string, error) {
	if opts.BaseOpenAPIJSONPath != "" {
		baseJSON, err := os.ReadFile(opts.BaseOpenAPIJSONPath)
		if err != nil {
			return "", err
		}
		if err := spec.UnmarshalJSON(baseJSON); err != nil {
			return "", fmt.Errorf("unmarshalling base: %w", err)
		}
	}

	if opts.BaseOpenAPIYAMLPath != "" {
		baseYAML, err := os.ReadFile(opts.BaseOpenAPIYAMLPath)
		if err != nil {
			return "", err
		}
		if err := spec.UnmarshalYAML(baseYAML); err != nil {
			return "", fmt.Errorf("unmarshalling base: %w", err)
		}
	}

	switch opts.Format {
	case "yaml":
		b, err := spec.MarshalYAML()
		if err != nil {
			return "", fmt.Errorf("marshalling: %w", err)
		}

		return string(b), nil
	case "json":
		b, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshalling: %w", err)
		}

		return string(b), nil
	default:
		return "", fmt.Errorf("unknown format: %s", opts.Format)
	}
}

func appendToSpec(opts Options, spec *openapi31.Spec, fd protoreflect.FileDescriptor) error {
	spec = gnostic.SpecWithFileAnnotations(spec, fd)
	components, err := fileToComponents(opts, fd)
	if err != nil {
		return err
	}
	for k, v := range components.Schemas {
		spec.Components.Schemas[k] = v
	}
	for k, v := range components.Responses {
		spec.Components.Responses[k] = v
	}
	for k, v := range components.Parameters {
		spec.Components.Parameters[k] = v
	}
	for k, v := range components.Examples {
		spec.Components.Examples[k] = v
	}
	for k, v := range components.RequestBodies {
		spec.Components.RequestBodies[k] = v
	}
	for k, v := range components.Headers {
		spec.Components.Headers[k] = v
	}
	for k, v := range components.SecuritySchemes {
		spec.Components.SecuritySchemes[k] = v
	}
	for k, v := range components.Links {
		spec.Components.Links[k] = v
	}
	for k, v := range components.Callbacks {
		spec.Components.Callbacks[k] = v
	}
	for k, v := range components.PathItems {
		spec.Components.PathItems[k] = v
	}

	pathItems, err := fileToPathItems(opts, fd)
	if err != nil {
		return err
	}
	for k, v := range pathItems {
		spec.Paths.MapOfPathItemValues[k] = v
	}
	spec.Tags = append(spec.Tags, fileToTags(fd)...)
	return nil
}
