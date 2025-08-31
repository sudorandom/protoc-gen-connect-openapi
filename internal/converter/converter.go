package converter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
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
	"google.golang.org/protobuf/types/dynamicpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

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

// Convert is the primary entrypoint for the protoc plugin. It takes a *pluginpb.CodeGeneratorRequest
// and returns a *pluginpb.CodeGeneratorResponse.
func Convert(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	opts, err := options.FromString(req.GetParameter())
	if err != nil {
		return nil, err
	}
	opts.Logger = slog.Default()
	return ConvertWithOptions(req, opts)
}

func ConvertWithOptions(req *pluginpb.CodeGeneratorRequest, opts options.Options) (*pluginpb.CodeGeneratorResponse, error) {
	if opts.Debug {
		opts.Logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level: slog.LevelDebug,
			}),
		)
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	annotator := &annotator{}
	if opts.MessageAnnotator == nil {
		opts.MessageAnnotator = annotator
	}
	if opts.FieldAnnotator == nil {
		opts.FieldAnnotator = annotator
	}
	if opts.FieldReferenceAnnotator == nil {
		opts.FieldReferenceAnnotator = annotator
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

	opts.ExtensionTypeResolver = dynamicpb.NewTypes(resolver)

	newSpec := func() (*v3.Document, error) {
		model := &v3.Document{}
		initializeDoc(opts, model)
		return model, nil
	}
	if len(opts.BaseOpenAPI) > 0 {
		newSpec = func() (*v3.Document, error) {
			document, err := libopenapi.NewDocument(opts.BaseOpenAPI)
			if err != nil {
				return &v3.Document{}, fmt.Errorf("unmarshalling base: %w", err)
			}
			v3Document, errs := document.BuildV3Model()
			if len(errs) > 0 {
				var merr error
				for _, err := range errs {
					merr = errors.Join(merr, err)
				}
				return &v3.Document{}, merr
			}
			model := &v3Document.Model
			initializeDoc(opts, model)
			return model, nil
		}
	}

	overrideComponents, err := getOverrideComponents(opts)
	if err != nil {
		return nil, err
	}

	spec, err := newSpec()
	if err != nil {
		return nil, err
	}
	outFiles := map[string]*v3.Document{}

	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			continue
		}

		opts.Logger.Debug("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := resolver.FindFileByPath(fileDesc.GetName())
		if err != nil {
			opts.Logger.Error("error loading file", slog.Any("error", err))
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

		if err := appendToSpec(opts, spec, fd); err != nil {
			return nil, err
		}

		if opts.Path == "" {
			name := fileDesc.GetName()
			filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".openapi." + opts.Format
			outFiles[filename] = spec
		}

		spec.Tags = mergeTags(spec.Tags)
		if overrideComponents != nil {
			util.AppendComponents(spec, overrideComponents)
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

	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	return &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &features,
		MinimumEdition:    proto.Int32(int32(descriptorpb.Edition_EDITION_PROTO2)),
		MaximumEdition:    proto.Int32(int32(descriptorpb.Edition_EDITION_2024)),
		File:              files,
	}, nil
}

func getOverrideComponents(opts options.Options) (*v3.Components, error) {
	if len(opts.OverrideOpenAPI) == 0 {
		return nil, nil
	}
	document, err := libopenapi.NewDocument(opts.OverrideOpenAPI)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling base: %w", err)
	}
	v3Document, errs := document.BuildV3Model()
	if len(errs) > 0 {
		var merr error
		for _, err := range errs {
			merr = errors.Join(merr, err)
		}
		return nil, merr
	}
	return v3Document.Model.Components, nil
}

func mergeTags(tags []*base.Tag) []*base.Tag {

	if len(tags) == 0 {
		return tags
	}

	res := make([]*base.Tag, 0, len(tags))
	found := make(map[string]*base.Tag)

	for _, tag := range tags {
		if found[tag.Name] == nil {
			found[tag.Name] = tag
			res = append(res, tag)
			continue
		}

		if tag.Description != "" {
			found[tag.Name].Description = tag.Description
		}

		if tag.ExternalDocs != nil {
			found[tag.Name].ExternalDocs = tag.ExternalDocs
		}

		if tag.Extensions != nil {
			found[tag.Name].Extensions = tag.Extensions
		}
	}

	return res
}

func specToFile(opts options.Options, spec *v3.Document) (string, error) {
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
	gnostic.SpecWithFileAnnotations(opts, spec, fd)

	components := &v3.Components{
		Schemas:         orderedmap.New[string, *base.SchemaProxy](),
		Responses:       orderedmap.New[string, *v3.Response](),
		Parameters:      orderedmap.New[string, *v3.Parameter](),
		Examples:        orderedmap.New[string, *base.Example](),
		RequestBodies:   orderedmap.New[string, *v3.RequestBody](),
		Headers:         orderedmap.New[string, *v3.Header](),
		SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
		Links:           orderedmap.New[string, *v3.Link](),
		Callbacks:       orderedmap.New[string, *v3.Callback](),
		Extensions:      orderedmap.New[string, *yaml.Node](),
	}

	// Only collect types from the root if TrimUnusedTypes is off
	if !opts.TrimUnusedTypes {
		// Files can have enums
		enums := fd.Enums()
		for i := 0; i < enums.Len(); i++ {
			AddEnumToSchema(opts, enums.Get(i), spec)
		}

		// Files can have messages
		messages := fd.Messages()
		for i := 0; i < messages.Len(); i++ {
			AddMessageSchemas(opts, messages.Get(i), spec)
		}
	}

	initializeDoc(opts, spec)
	appendServiceDocs(opts, spec, fd)
	initializeComponents(components)
	util.AppendComponents(spec, components)

	if err := addPathItemsFromFile(opts, fd, spec); err != nil {
		return err
	}
	spec.Tags = append(spec.Tags, fileToTags(opts, fd)...)

	// Sort
	orderedmap.SortAlpha(spec.Paths.PathItems)
	orderedmap.SortAlpha(spec.Components.Schemas)
	orderedmap.SortAlpha(spec.Components.Responses)
	orderedmap.SortAlpha(spec.Components.Parameters)
	orderedmap.SortAlpha(spec.Components.Examples)
	orderedmap.SortAlpha(spec.Components.RequestBodies)
	orderedmap.SortAlpha(spec.Components.Headers)
	orderedmap.SortAlpha(spec.Components.SecuritySchemes)
	orderedmap.SortAlpha(spec.Components.Links)
	orderedmap.SortAlpha(spec.Components.Callbacks)
	orderedmap.SortAlpha(spec.Components.PathItems)
	orderedmap.SortAlpha(spec.Components.Extensions)

	return nil
}

func appendServiceDocs(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) {
	if !opts.WithServiceDescriptions {
		return
	}
	var builder strings.Builder
	if spec.Info.Description != "" {
		builder.WriteString(spec.Info.Description)
		builder.WriteString("\n\n")
	}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}

		builder.WriteString("## ")
		builder.WriteString(string(service.FullName()))
		builder.WriteString("\n\n")

		loc := fd.SourceLocations().ByDescriptor(service)
		serviceComments := util.FormatComments(loc)
		if serviceComments != "" {
			builder.WriteString(serviceComments)
			builder.WriteString("\n\n")
		}
	}

	spec.Info.Description = strings.TrimSpace(builder.String())
}

func initializeDoc(opts options.Options, doc *v3.Document) {
	opts.Logger.Debug("initializeDoc")
	if doc.Version == "" {
		doc.Version = "3.1.0"
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
	if doc.Security == nil {
		doc.Security = []*base.SecurityRequirement{}
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
	if doc.Components == nil {
		doc.Components = &v3.Components{}
	}
	initializeComponents(doc.Components)
}

func initializeComponents(components *v3.Components) {
	if components.Schemas == nil {
		components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}
	if components.Responses == nil {
		components.Responses = orderedmap.New[string, *v3.Response]()
	}
	if components.Parameters == nil {
		components.Parameters = orderedmap.New[string, *v3.Parameter]()
	}
	if components.Examples == nil {
		components.Examples = orderedmap.New[string, *base.Example]()
	}
	if components.RequestBodies == nil {
		components.RequestBodies = orderedmap.New[string, *v3.RequestBody]()
	}
	if components.Headers == nil {
		components.Headers = orderedmap.New[string, *v3.Header]()
	}
	if components.SecuritySchemes == nil {
		components.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}
	if components.Links == nil {
		components.Links = orderedmap.New[string, *v3.Link]()
	}
	if components.Callbacks == nil {
		components.Callbacks = orderedmap.New[string, *v3.Callback]()
	}
	if components.Extensions == nil {
		components.Extensions = orderedmap.New[string, *yaml.Node]()
	}
}
