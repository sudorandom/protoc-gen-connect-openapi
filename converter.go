package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Options struct {
	// Format can be either "yaml" or "json"
	Format  string
	Version string
}

func parseOptions(s string) (Options, error) {
	opts := Options{
		Version: "v1.0.0",
		Format:  "yaml",
	}

	for _, param := range strings.Split(s, ",") {
		switch param {
		case "":
		case "format_yaml":
			opts.Format = "yaml"
		case "format_json":
			opts.Format = "json"
		default:
			if strings.HasPrefix(param, "version=") {
				opts.Version = param[8:]
			} else {
				return opts, fmt.Errorf("invalid parameter: %s", param)
			}
		}
	}
	return opts, nil
}

func ConvertFrom(rd io.Reader) (*plugin.CodeGeneratorResponse, error) {
	input, err := io.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	req := &plugin.CodeGeneratorRequest{}
	err = proto.Unmarshal(input, req)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal input: %w", err)
	}

	return Convert(req)
}

func formatTypeRef(t string) string {
	return strings.TrimPrefix(t, ".")
}

func Convert(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	opts, err := parseOptions(req.GetParameter())
	if err != nil {
		return nil, err
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

	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			slog.Debug("skip generating file because it wasn't requested", slog.String("name", fileDesc.GetName()))
			continue
		}

		slog.Info("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := protodesc.NewFile(fileDesc, resolver)
		if err != nil {
			slog.Error("error loading file", slog.Any("error", err))
			return nil, err
		}

		spec := openapi31.Spec{Openapi: "3.1.0"}
		spec.SetTitle(string(fd.FullName()))
		spec.SetDescription(formatComments(fd.SourceLocations().ByDescriptor(fd)))
		spec.SetVersion(opts.Version)

		// Add all messages as top-level types
		components := openapi31.Components{}
		state := &State{}
		rootSchema := fileToSchema(state, fd)
		for _, item := range rootSchema.Items.SchemaArray {
			if item.TypeObject == nil {
				continue
			}
			m, err := item.ToSimpleMap()
			if err != nil {
				return nil, err
			}
			components.WithSchemasItem(*item.TypeObject.ID, m)
		}

		// Add our own type for errors
		reflector := jsonschema.Reflector{}
		connectError, err := reflector.Reflect(ConnectError{})
		if err != nil {
			return nil, err
		}
		connectError.WithID("connect.error")

		components.WithSchemasItem(*connectError.ID, map[string]interface{}{
			"$id":         connectError.ID,
			"description": connectError.Description,
			"properties":  connectError.Properties,
			"title":       connectError.Title,
			"type":        connectError.Type,
		})

		components.WithResponsesItem("connect.error", openapi31.ResponseOrReference{
			Reference: &openapi31.Reference{
				Ref: "#/components/schemas/connect.error",
			},
		})

		spec.WithComponents(components)

		// Add "path items", which are service methods
		items := map[string]openapi31.PathItem{}
		tags := []openapi31.Tag{}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			loc := fd.SourceLocations().ByDescriptor(service)
			description := formatComments(loc)

			tags = append(tags, openapi31.Tag{
				Name:        string(service.FullName()),
				Description: &description,
			})

			methods := service.Methods()
			for j := 0; j < services.Len(); j++ {
				method := methods.Get(j)
				op := &openapi31.Operation{}
				op.WithTags(string(service.FullName()))
				loc := fd.SourceLocations().ByDescriptor(method)
				op.WithDescription(formatComments(loc))

				// Request Body
				item := openapi31.PathItem{}
				op.WithRequestBody(openapi31.RequestBodyOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/requestBodies/" + formatTypeRef(string(method.Input().FullName())),
					},
				})
				trueVar := true
				spec.Components.WithRequestBodiesItem(formatTypeRef(string(method.Input().FullName())),
					openapi31.RequestBodyOrReference{
						RequestBody: &openapi31.RequestBody{
							Description: new(string),
							Content: map[string]openapi31.MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"$ref": "#/components/schemas/" + formatTypeRef(string(method.Input().FullName())),
									},
								},
							},
							Required:      &trueVar,
							MapOfAnything: map[string]interface{}{},
						},
					},
				)

				// Responses
				op.WithResponses(openapi31.Responses{
					Default: &openapi31.ResponseOrReference{
						Response: &openapi31.Response{
							Content: map[string]openapi31.MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"$ref": "#/components/responses/connect.error",
									},
								},
							},
						},
					},
					MapOfResponseOrReferenceValues: map[string]openapi31.ResponseOrReference{
						"200": {
							Response: &openapi31.Response{
								Description: description,
								Content: map[string]openapi31.MediaType{
									"application/json": {
										Schema: map[string]interface{}{
											"$ref": "#/components/responses/" + formatTypeRef(string(method.Output().FullName())),
										},
									},
								},
							},
						},
					},
				})

				spec.Components.WithResponsesItem(formatTypeRef(string(method.Output().FullName())),
					openapi31.ResponseOrReference{
						Reference: &openapi31.Reference{
							Ref: "#/components/schemas/" + formatTypeRef(string(method.Output().FullName())),
						},
					},
				)

				options := method.Options().(*descriptorpb.MethodOptions)
				if options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS {
					item.Get = op
				} else {
					item.Post = op
				}
				items["/"+string(service.FullName())+"/"+string(method.Name())] = item
			}
		}
		spec.WithPaths(openapi31.Paths{MapOfPathItemValues: items})
		spec.WithTags(tags...)

		switch opts.Format {
		case "yaml":
			b, err := spec.MarshalYAML()
			if err != nil {
				return nil, err
			}

			content := string(b)
			name := fileDesc.GetName()
			filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".openapi.yaml"

			files = append(files, &pluginpb.CodeGeneratorResponse_File{
				Name:              &filename,
				Content:           &content,
				GeneratedCodeInfo: &descriptorpb.GeneratedCodeInfo{},
			})
		case "json":
			b, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return nil, err
			}

			content := string(b)
			name := fileDesc.GetName()
			filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".openapi.json"

			files = append(files, &pluginpb.CodeGeneratorResponse_File{
				Name:              &filename,
				Content:           &content,
				GeneratedCodeInfo: &descriptorpb.GeneratedCodeInfo{},
			})
		}
	}

	return &plugin.CodeGeneratorResponse{
		File: files,
	}, nil
}

type State struct {
	CurrentFile      protoreflect.FileDescriptor
	ExternalMessages []protoreflect.MessageDescriptor
	ExternalEnums    []protoreflect.EnumDescriptor
}

func enumToSchema(state *State, tt protoreflect.EnumDescriptor) *jsonschema.Schema {
	slog.Info("enumToSchema", slog.Any("descriptor", tt.FullName()))
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithType(jsonschema.String.Type())
	children := []interface{}{}
	values := tt.Values()
	for i := 0; i < values.Len(); i++ {
		value := values.Get(i)
		children = append(children, string(value.Name()))
		children = append(children, value.Number())
	}
	s.WithEnum(children)
	return s
}

func messageToSchema(state *State, tt protoreflect.MessageDescriptor) *jsonschema.Schema {
	slog.Info("messageToSchema", slog.Any("descriptor", tt.FullName()))
	if isWellKnown(tt) {
		return wellKnownToSchema(tt)
	}
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithType(jsonschema.Object.Type())

	fields := tt.Fields()
	children := make(map[string]jsonschema.SchemaOrBool, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		child := fieldToSchema(state, field)
		children[field.JSONName()] = jsonschema.SchemaOrBool{TypeObject: child}
	}

	s.WithProperties(children)
	return s
}

func fieldToSchema(state *State, tt protoreflect.FieldDescriptor) *jsonschema.Schema {
	slog.Info("fieldToSchema", slog.Any("descriptor", tt.FullName()))
	s := &jsonschema.Schema{}

	switch tt.Kind() {
	case protoreflect.BoolKind:
		s.WithType(jsonschema.Boolean.Type())
	case protoreflect.EnumKind:
		if tt.Enum().ParentFile().Path() != state.CurrentFile.Path() {
			state.ExternalEnums = append(state.ExternalEnums, tt.Enum())
		}
		s.WithRef("#/components/schemas/" + string(tt.Enum().FullName()))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind:
		s.WithType(jsonschema.Integer.Type())
	case protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		s.WithType(jsonschema.Integer.Type())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.FloatKind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.DoubleKind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.StringKind:
		s.WithType(jsonschema.String.Type())
	case protoreflect.BytesKind:
		s.WithType(jsonschema.String.Type())
	case protoreflect.MessageKind:
		if tt.Message().ParentFile().Path() != state.CurrentFile.Path() {
			state.ExternalMessages = append(state.ExternalMessages, tt.Message())
		}
		s.WithRef("#/components/schemas/" + string(tt.Message().FullName()))
		s.WithType(jsonschema.Object.Type())
	}

	// Handle maps
	if tt.IsMap() {
		s.AdditionalProperties = &jsonschema.SchemaOrBool{TypeObject: fieldToSchema(state, tt.MapValue())}
		s.WithType(jsonschema.Object.Type())
		s.Ref = nil
	}

	// Handle Lists
	if tt.IsList() {
		wrapped := s
		s = &jsonschema.Schema{}
		s.WithType(jsonschema.Array.Type())
		s.WithItems(jsonschema.Items{SchemaArray: []jsonschema.SchemaOrBool{{TypeObject: wrapped}}})
	}

	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	return s
}

func fileToSchema(state *State, tt protoreflect.FileDescriptor) *jsonschema.Schema {
	slog.Info("fileOfToSchema", slog.Any("descriptor", tt.FullName()))
	state.CurrentFile = tt
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	children := []jsonschema.SchemaOrBool{}
	enums := tt.Enums()
	for i := 0; i < enums.Len(); i++ {
		child := enumToSchema(state, enums.Get(i))
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
	}
	messages := tt.Messages()
	for i := 0; i < messages.Len(); i++ {
		message := messages.Get(i)
		child := messageToSchema(state, message)
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})

		// messages can also have enums defined in them. This is handled by adding them to the root
		// schema with a fully-qualified path
		enums := message.Enums()
		for i := 0; i < enums.Len(); i++ {
			child := enumToSchema(state, enums.Get(i))
			children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
		}
	}

	for _, child := range resolveExternalDescriptors(state) {
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
	}

	s.WithItems(jsonschema.Items{SchemaArray: children})
	return s
}

func resolveExternalDescriptors(state *State) []*jsonschema.Schema {
	if len(state.ExternalEnums) == 0 && len(state.ExternalMessages) == 0 {
		return []*jsonschema.Schema{}
	}

	children := []*jsonschema.Schema{}
	for _, enum := range state.ExternalEnums {
		childState := &State{
			CurrentFile: enum.ParentFile(),
		}
		children = append(children, enumToSchema(childState, enum))
		children = append(children, resolveExternalDescriptors(childState)...)
	}

	for _, message := range state.ExternalMessages {
		childState := &State{
			CurrentFile: message.ParentFile(),
		}
		children = append(children, messageToSchema(childState, message))
		children = append(children, resolveExternalDescriptors(childState)...)
	}

	return children
}

type ConnectError struct {
	Code    string `json:"code" example:"CodeNotFound" enum:"CodeCanceled,CodeUnknown,CodeInvalidArgument,CodeDeadlineExceeded,CodeNotFound,CodeAlreadyExists,CodePermissionDenied,CodeResourceExhausted,CodeFailedPrecondition,CodeAborted,CodeOutOfRange,CodeInternal,CodeUnavailable,CodeDataLoss,CodeUnauthenticated"`
	Message string `json:"message,omitempty"`
}

func formatComments(loc protoreflect.SourceLocation) string {
	var builder strings.Builder
	if loc.LeadingComments != "" {
		builder.WriteString(strings.TrimSpace(loc.LeadingComments))
		builder.WriteString(" ")
	}
	if loc.TrailingComments != "" {
		builder.WriteString(strings.TrimSpace(loc.TrailingComments))
		builder.WriteString(" ")
	}
	return strings.TrimSpace(builder.String())
}
