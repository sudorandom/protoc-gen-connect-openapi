package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"reflect"
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

// TODO: Option to set API version
// TODO: Handle well-known types
// TODO: Additional string types
// TODO: Extra credit: protovalidate constraints

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
	files := []*pluginpb.CodeGeneratorResponse_File{}
	genFiles := make(map[string]struct{}, len(req.FileToGenerate))
	for _, file := range req.FileToGenerate {
		genFiles[file] = struct{}{}
	}
	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			slog.Info("skip generating file because it wasn't requested", slog.String("name", fileDesc.GetName()))
			continue
		}

		slog.Info("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := protodesc.NewFile(fileDesc, nil)
		if err != nil {
			return nil, err
		}

		spec := openapi31.Spec{Openapi: "3.1.0"}
		spec.SetTitle(string(fd.FullName()))

		spec.SetDescription("")

		// TODO: This should come in from CLI arguments, this data isn't contained in proto files
		spec.SetVersion("v1.0.0")

		// Add all messages as top-level types
		components := openapi31.Components{}
		rootSchema := resolveJsonSchema(fd)
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
		components.WithSchemasItem("connect.error", map[string]interface{}{
			"properties": connectError.Properties,
			"type":       connectError.Type,
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

				item := openapi31.PathItem{
					Parameters: []openapi31.ParameterOrReference{
						{
							Reference: &openapi31.Reference{
								Ref: "#/components/schemas/" + formatTypeRef(string(method.Output().FullName())),
							},
						},
					},
				}

				op.WithResponses(openapi31.Responses{
					Default: &openapi31.ResponseOrReference{
						Response: &openapi31.Response{
							Content: map[string]openapi31.MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"$ref": "#/components/schemas/connect.error",
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
											"$ref": "#/components/schemas/" + formatTypeRef(string(method.Input().FullName())),
										},
									},
								},
							},
						},
					},
				})

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

	return &plugin.CodeGeneratorResponse{
		File: files,
	}, nil
}

func resolveJsonSchema(t protoreflect.Descriptor) *jsonschema.Schema {
	slog.Info("processSchemaItem", slog.Any("descriptor", t.FullName()), slog.Any("type", reflect.TypeOf(t).String()))
	switch tt := t.(type) {
	case protoreflect.EnumDescriptor:
		s := &jsonschema.Schema{}
		s.WithID(string(t.FullName()))
		s.WithTitle(string(tt.Name()))
		s.WithDescription(formatComments(t.ParentFile().SourceLocations().ByDescriptor(t)))
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
	case protoreflect.EnumValueDescriptor:
	case protoreflect.MessageDescriptor:
		s := &jsonschema.Schema{}
		s.WithID(string(t.FullName()))
		s.WithTitle(string(tt.Name()))
		s.WithDescription(formatComments(t.ParentFile().SourceLocations().ByDescriptor(t)))
		s.WithType(jsonschema.Object.Type())

		fields := tt.Fields()
		children := make(map[string]jsonschema.SchemaOrBool, fields.Len())
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			child := resolveJsonSchema(field)
			children[field.JSONName()] = jsonschema.SchemaOrBool{TypeObject: child}
		}
		s.WithProperties(children)
		return s
	case protoreflect.FieldDescriptor:
		s := &jsonschema.Schema{}
		s.WithID(string(tt.FullName()))
		s.WithTitle(string(tt.Name()))
		s.WithDescription(formatComments(t.ParentFile().SourceLocations().ByDescriptor(t)))
		if tt.IsMap() {
			s.AdditionalProperties = &jsonschema.SchemaOrBool{TypeObject: resolveJsonSchema(tt.MapValue())}
		}
		switch tt.Kind() {
		case protoreflect.BoolKind:
			s.WithType(jsonschema.Boolean.Type())
		case protoreflect.EnumKind:
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
			s.WithRef("#/components/schemas/" + string(tt.Message().FullName()))
			s.WithType(jsonschema.Object.Type())
		}

		// Handle maps
		// tt.MapKey()
		// tt.MapValue()

		// Handle Lists
		// if tt.IsList() {
		// 	s.WithType(jsonschema.Object.Type())
		// }
		return s
	case protoreflect.OneofDescriptor:
		s := &jsonschema.Schema{}
		s.WithID(string(tt.FullName()))
		s.WithTitle(string(tt.Name()))
		s.WithDescription(formatComments(t.ParentFile().SourceLocations().ByDescriptor(t)))
		children := []jsonschema.SchemaOrBool{}
		fields := tt.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			children = append(children, jsonschema.SchemaOrBool{TypeObject: resolveJsonSchema(field)})
		}
		s.WithOneOf(children...)
		return s
	case protoreflect.FileDescriptor:
		s := &jsonschema.Schema{}
		s.WithID(string(t.FullName()))
		s.WithTitle(string(tt.Name()))
		s.WithDescription(formatComments(t.ParentFile().SourceLocations().ByDescriptor(t)))
		children := []jsonschema.SchemaOrBool{}
		enums := tt.Enums()
		for i := 0; i < enums.Len(); i++ {
			child := resolveJsonSchema(enums.Get(i))
			children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
		}
		messages := tt.Messages()
		for i := 0; i < messages.Len(); i++ {
			message := messages.Get(i)
			child := resolveJsonSchema(message)
			children = append(children, jsonschema.SchemaOrBool{TypeObject: child})

			// messages can also have enums defined in them. This is handled by adding them to the root
			// schema with a fully-qualified path
			enums := message.Enums()
			for i := 0; i < enums.Len(); i++ {
				child := resolveJsonSchema(enums.Get(i))
				children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
			}
		}
		s.WithItems(jsonschema.Items{SchemaArray: children})
		return s

	// We don't use these here
	case protoreflect.ServiceDescriptor:
	case protoreflect.MethodDescriptor:
	}

	return nil
}

// TODO: Add more annotations for examples
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
