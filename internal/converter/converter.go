package converter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
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
	return ConvertWithOptions(req, opts)
}

func ConvertWithOptions(req *pluginpb.CodeGeneratorRequest, opts options.Options) (*pluginpb.CodeGeneratorResponse, error) {
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

	newSpec := func() (*v3.Document, error) {
		model := &v3.Document{}
		initializeDoc(model)
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
			initializeDoc(model)
			return model, nil
		}
	}

	spec, err := newSpec()
	if err != nil {
		return nil, err
	}
	outFiles := map[string]*v3.Document{}

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

		if err := appendToSpec(opts, spec, fd); err != nil {
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

	features := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	return &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &features,
		MinimumEdition:    proto.Int32(int32(descriptor.Edition_EDITION_PROTO2)),
		MaximumEdition:    proto.Int32(int32(descriptor.Edition_EDITION_2024)),
		File:              files,
	}, nil
}

func specToFile(opts options.Options, spec *v3.Document) (string, error) {
	if opts.TrimUnusedTypes {
		if err := trimUnusedTypes(spec); err != nil {
			return "", err
		}
	}
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

func TrimUnusedTypes(spec *v3.Document) error {
	return trimUnusedTypes(spec)
}

func trimUnusedTypes(spec *v3.Document) error {
	slog.Debug("trimming unused types")
	
	// Get all references from the document
	b, err := spec.Render()
	if err != nil {
		return err
	}
	doc, err := libopenapi.NewDocument(b)
	if err != nil {
		return err
	}
	model, errs := doc.BuildV3Model()
	if errs != nil {
		return errors.Join(errs...)
	}
	model.Index.BuildIndex()
	references := model.Model.Rolodex.GetRootIndex().GetAllReferences()

	// Process each schema
	deletedKeys := make(map[string]bool)
	for pair := spec.Components.Schemas.First(); pair != nil; pair = pair.Next() {
		key := pair.Key()
		ref := fmt.Sprintf("#/components/schemas/%s", key)
		
		// Skip if already used in the root document
		if _, usedInRootIndex := references[ref]; usedInRootIndex {
			continue
		}

		// Check if referenced by other components
		usedInComponents := false
		for otherPair := spec.Components.Schemas.First(); otherPair != nil; otherPair = otherPair.Next() {
			if otherPair.Key() == key || deletedKeys[otherPair.Key()] {
				continue
			}
			
			otherRefs := getReferencesInSchema(otherPair.Value())
			if _, ok := otherRefs[ref]; ok {
				usedInComponents = true
				break
			}
		}

		if !usedInComponents {
			slog.Debug("trimming unused type", "name", key)
			trimChildSchemas(spec.Components.Schemas, key)
			spec.Components.Schemas.Delete(key)
			deletedKeys[key] = true
		}
	}

	return nil
}

// recursively delete all schemas that are referenced by the given schema, but only if the child is also unused
func trimChildSchemas(schemas *orderedmap.Map[string, *base.SchemaProxy], key string) {
    parentSchema, ok := schemas.Get(key)
    if !ok {
        return
    }
    
    // Get all references in the schema being deleted
    childReferences := getReferencesInSchema(parentSchema)
    
    // For each child reference, check if it's used by other schemas
    for childRef := range childReferences {
        // Extract the schema name from the reference (e.g., "#/components/schemas/Pet" -> "Pet")
        childKey := strings.TrimPrefix(childRef, "#/components/schemas/")
        
        // Skip if this schema doesn't exist
        _, ok := schemas.Get(childKey)
        if !ok {
            continue
        }
        
        // Check if this child schema is referenced by any other schemas
        isUsedElsewhere := false
        for pair := schemas.First(); pair != nil; pair = pair.Next() {
            // Skip the parent schema we're currently processing
            if pair.Key() == key {
                continue
            }
            
            otherSchemaRefs := getReferencesInSchema(pair.Value())
            if _, ok := otherSchemaRefs[childRef]; ok {
                isUsedElsewhere = true
                break
            }
        }
        
        // Only recursively delete if this schema isn't used elsewhere
        if !isUsedElsewhere {
            trimChildSchemas(schemas, childKey)
            schemas.Delete(childKey)
        }
    }
}

func getReferencesInSchema(schProxy *base.SchemaProxy) map[string]string {
	references := make(map[string]string)
	if schProxy == nil {
		return references
	}

	// Handle direct references
	if schProxy.IsReference() {
		ref := schProxy.GetReference()
		references[ref] = ref
		return references
	}

	schema, err := schProxy.BuildSchema()
	if err != nil {
		slog.Error("error building schema", slog.Any("error", err))
		return references
	}

	// Helper function to process a schema proxy
	processProxy := func(proxy *base.SchemaProxy) {
		if proxy == nil {
			return
		}
		for k, v := range getReferencesInSchema(proxy) {
			references[k] = v
		}
	}

	// Process properties
	for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
		processProxy(pair.Value())
	}

	// Process array items
	if schema.Items != nil {
		processProxy(schema.Items.A)
	}

	// Process additional properties
	if schema.AdditionalProperties != nil {
		processProxy(schema.AdditionalProperties.A)
	}

	// Process composition schemas
	for _, s := range schema.AllOf {
		processProxy(s)
	}
	for _, s := range schema.OneOf {
		processProxy(s)
	}
	for _, s := range schema.AnyOf {
		processProxy(s)
	}
	if schema.Not != nil {
		processProxy(schema.Not)
	}

	return references
}

func appendToSpec(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) error {
	gnostic.SpecWithFileAnnotations(spec, fd)
	components, err := fileToComponents(opts, fd)
	if err != nil {
		return err
	}
	initializeDoc(spec)
	initializeComponents(components)
	util.AppendComponents(spec, components)

	if err := addPathItemsFromFile(opts, fd, spec.Paths); err != nil {
		return err
	}
	spec.Tags = append(spec.Tags, fileToTags(fd)...)
	return nil
}

func initializeDoc(doc *v3.Document) {
	slog.Debug("initializeDoc")
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
