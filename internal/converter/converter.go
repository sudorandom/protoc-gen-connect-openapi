package converter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/pb33f/libopenapi"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func parseOptions(s string) (options.Options, error) {
	opts := options.Options{
		Format:       "yaml",
		ContentTypes: map[string]struct{}{},
	}

	supportedProtocols := map[string]struct{}{}
	for _, proto := range options.Protocols {
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
			case ".yaml", ".yml", ".json":
				opts.BaseOpenAPIPath = basePath
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

	if opts.Debug {
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level: slog.LevelDebug,
			}),
		))
	}

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

	newSpec := func() (v3.Document, error) {
		return initializeDoc(v3.Document{}), nil
	}
	if opts.BaseOpenAPIPath != "" {
		newSpec = func() (v3.Document, error) {
			base, err := os.ReadFile(opts.BaseOpenAPIPath)
			if err != nil {
				return v3.Document{}, err
			}

			document, err := libopenapi.NewDocument(base)
			if err != nil {
				return v3.Document{}, fmt.Errorf("unmarshalling base: %w", err)
			}
			v3Document, errs := document.BuildV3Model()
			if len(errs) > 0 {
				var merr error
				for _, err := range errs {
					merr = errors.Join(merr, err)
				}
				return v3.Document{}, merr
			}
			return initializeDoc(v3Document.Model), nil
		}
	}

	spec, err := newSpec()
	if err != nil {
		return nil, err
	}
	outFiles := map[string]v3.Document{}

	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			slog.Debug("skip generating file because it wasn't requested", slog.String("name", fileDesc.GetName()))
			continue
		}

		slog.Debug("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := resolver.FindFileByPath(fileDesc.GetName())
		if err != nil {
			slog.Error("error loading file", slog.Any("error", err))
			return nil, err
		}

		// Create a per-file openapi spec if we're not merging all into one
		if opts.Path == "" {
			spec, err = newSpec()
			if err != nil {
				return nil, err
			}
			spec.Info.Title = string(fd.FullName())
			spec.Info.Description = util.FormatComments(fd.SourceLocations().ByDescriptor(fd))
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

func specToFile(opts options.Options, spec v3.Document) (string, error) {
	switch opts.Format {
	case "yaml":
		return string(spec.RenderWithIndention(2)), nil
	case "json":
		b, err := spec.RenderJSON("  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("unknown format: %s", opts.Format)
	}
}

func appendToSpec(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) error {
	spec = gnostic.SpecWithFileAnnotations(spec, fd)
	components, err := fileToComponents(opts, fd)
	if err != nil {
		return err
	}
	for pair := components.Schemas.First(); pair != nil; pair = pair.Next() {
		spec.Components.Schemas.Set(pair.Key(), pair.Value())
	}
	for pair := components.Responses.First(); pair != nil; pair = pair.Next() {
		spec.Components.Responses.Set(pair.Key(), pair.Value())
	}
	for pair := components.Parameters.First(); pair != nil; pair = pair.Next() {
		spec.Components.Parameters.Set(pair.Key(), pair.Value())
	}
	for pair := components.Examples.First(); pair != nil; pair = pair.Next() {
		spec.Components.Examples.Set(pair.Key(), pair.Value())
	}
	for pair := components.RequestBodies.First(); pair != nil; pair = pair.Next() {
		spec.Components.RequestBodies.Set(pair.Key(), pair.Value())
	}
	for pair := components.Headers.First(); pair != nil; pair = pair.Next() {
		spec.Components.Headers.Set(pair.Key(), pair.Value())
	}
	for pair := components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
		spec.Components.SecuritySchemes.Set(pair.Key(), pair.Value())
	}
	for pair := components.Links.First(); pair != nil; pair = pair.Next() {
		spec.Components.Links.Set(pair.Key(), pair.Value())
	}
	for pair := components.Callbacks.First(); pair != nil; pair = pair.Next() {
		spec.Components.Callbacks.Set(pair.Key(), pair.Value())
	}

	pathItems, err := fileToPathItems(opts, fd)
	if err != nil {
		return err
	}
	for pair := pathItems.First(); pair != nil; pair = pair.Next() {
		spec.Paths.PathItems.Set(pair.Key(), pair.Value())
	}
	spec.Tags = append(spec.Tags, fileToTags(fd)...)
	return nil
}

func initializeDoc(doc v3.Document) v3.Document {
	if doc.Version == "" {
		doc.Version = "3.1.0"
	}
	if doc.Paths == nil {
		doc.Paths = &v3.Paths{}
	}
	if doc.Paths.PathItems == nil {
		doc.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Paths.Extensions == nil {
		doc.Paths.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Info == nil {
		doc.Info = &base.Info{}
	}
	if doc.Paths == nil {
		doc.Paths = &v3.Paths{}
	}
	if doc.Paths.PathItems == nil {
		doc.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Paths.Extensions == nil {
		doc.Paths.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Components == nil {
		doc.Components = &v3.Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}
	if doc.Components.Responses == nil {
		doc.Components.Responses = orderedmap.New[string, *v3.Response]()
	}
	if doc.Components.Parameters == nil {
		doc.Components.Parameters = orderedmap.New[string, *v3.Parameter]()
	}
	if doc.Components.Examples == nil {
		doc.Components.Examples = orderedmap.New[string, *base.Example]()
	}
	if doc.Components.RequestBodies == nil {
		doc.Components.RequestBodies = orderedmap.New[string, *v3.RequestBody]()
	}
	if doc.Components.Headers == nil {
		doc.Components.Headers = orderedmap.New[string, *v3.Header]()
	}
	if doc.Components.SecuritySchemes == nil {
		doc.Components.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}
	if doc.Components.Links == nil {
		doc.Components.Links = orderedmap.New[string, *v3.Link]()
	}
	if doc.Components.Callbacks == nil {
		doc.Components.Callbacks = orderedmap.New[string, *v3.Callback]()
	}
	if doc.Components.Extensions == nil {
		doc.Components.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Security == nil {
		doc.Security = []*base.SecurityRequirement{}
	}
	if doc.ExternalDocs == nil {
		doc.ExternalDocs = &base.ExternalDoc{}
	}
	if doc.Extensions == nil {
		doc.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Webhooks == nil {
		doc.Webhooks = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Index == nil {
		doc.Index = &index.SpecIndex{}
	}
	if doc.Rolodex == nil {
		doc.Rolodex = &index.Rolodex{}
	}

	return doc
}
