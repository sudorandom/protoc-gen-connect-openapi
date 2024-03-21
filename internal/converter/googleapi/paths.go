package googleapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func MakePathItems(md protoreflect.MethodDescriptor) map[string]openapi31.PathItem {
	opts := md.Options()
	if !proto.HasExtension(opts, annotations.E_Http) {
		return nil
	}
	rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return nil
	}
	return httpRuleToPathMap(md, rule)
}

func httpRuleToPathMap(md protoreflect.MethodDescriptor, rule *annotations.HttpRule) map[string]openapi31.PathItem {
	var method, template string
	switch pattern := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		method, template = http.MethodGet, pattern.Get
	case *annotations.HttpRule_Put:
		method, template = http.MethodPut, pattern.Put
	case *annotations.HttpRule_Post:
		method, template = http.MethodPost, pattern.Post
	case *annotations.HttpRule_Delete:
		method, template = http.MethodDelete, pattern.Delete
	case *annotations.HttpRule_Patch:
		method, template = http.MethodPatch, pattern.Patch
	case *annotations.HttpRule_Custom:
		method, template = pattern.Custom.GetKind(), pattern.Custom.GetPath()
	default:
		slog.Warn("invalid type of pattern for HTTP rule", slog.Any("pattern", pattern))
		return nil
	}
	if method == "" {
		slog.Warn("invalid HTTP rule: method is blank", slog.Any("method", md))
		return nil
	}
	if template == "" {
		slog.Warn("invalid HTTP rule: path template is blank", slog.Any("method", md))
		return nil
	}

	tokens, err := RunPathPatternLexer(template)
	if err != nil {
		slog.Warn("unable to parse template pattern", slog.Any("error", err), slog.String("template", template))
		return nil
	}

	paths := map[string]openapi31.PathItem{}
	pathItem := openapi31.PathItem{}

	fd := md.ParentFile()
	service := md.Parent().(protoreflect.ServiceDescriptor)
	op := &openapi31.Operation{}
	op.WithTags(string(service.FullName()))
	loc := fd.SourceLocations().ByDescriptor(md)
	op.WithDescription(util.FormatComments(loc))

	parameters := []openapi31.ParameterOrReference{}
	switch rule.Body {
	case "":
		fields := md.Input().Fields()
		for i := 0; i < fields.Len(); i++ {
			loc := fd.SourceLocations().ByDescriptor(md)
			desc := util.FormatComments(loc)
			field := fields.Get(i)
			parameters = append(parameters, openapi31.ParameterOrReference{
				Parameter: &openapi31.Parameter{
					Name:        field.JSONName(),
					In:          "query",
					Description: &desc,
					Schema:      schemaToMap(util.FieldToSchema(nil, field)),
				},
			})
		}
	case "*":
		id := util.FormatTypeRef(string(md.FullName() + "." + md.Input().FullName()))
		op.WithRequestBody(openapi31.RequestBodyOrReference{
			Reference: &openapi31.Reference{Ref: "#/components/requestBodies/" + id},
		})
	default:
		fields := md.Input().Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if field.JSONName() != rule.Body {
				continue
			}
			loc := fd.SourceLocations().ByDescriptor(md)
			desc := util.FormatComments(loc)
			op.WithRequestBody(openapi31.RequestBodyOrReference{
				RequestBody: &openapi31.RequestBody{
					Description: &desc,
				},
			})
		}
	}

	for _, param := range partsToParameter(tokens) {
		field := resolveField(md.Input(), param)
		if field != nil {
			loc := fd.SourceLocations().ByDescriptor(field)
			desc := util.FormatComments(loc)
			parameters = append(parameters, openapi31.ParameterOrReference{
				Parameter: &openapi31.Parameter{
					Name:        param,
					In:          "path",
					Description: &desc,
					Schema:      schemaToMap(util.FieldToSchema(nil, field)),
				},
			})
		}
	}
	op.WithParameters(parameters...)

	// Responses
	responses := openapi31.Responses{
		Default: &openapi31.ResponseOrReference{
			Reference: &openapi31.Reference{Ref: "#/components/responses/connect.error"},
		},
	}
	if !util.IsEmpty(md.Output()) {
		id := util.FormatTypeRef(string(md.FullName() + "." + md.Output().FullName()))
		responses.WithMapOfResponseOrReferenceValuesItem("200", openapi31.ResponseOrReference{
			Reference: &openapi31.Reference{Ref: "#/components/responses/" + id},
		})
	}
	op.WithResponses(responses)

	switch method {
	case http.MethodGet:
		pathItem.Get = op
	case http.MethodPut:
		pathItem.Put = op
	case http.MethodPost:
		pathItem.Post = op
	case http.MethodDelete:
		pathItem.Delete = op
	case http.MethodPatch:
		pathItem.Patch = op
	default:
		pathItem.MapOfAnything[method] = op
	}
	paths[partsToOpenAPIPath(tokens)] = pathItem

	for _, binding := range rule.AdditionalBindings {
		for k, v := range httpRuleToPathMap(md, binding) {
			paths[k] = v
		}
	}
	return paths
}

func resolveField(md protoreflect.MessageDescriptor, param string) protoreflect.FieldDescriptor {
	var current protoreflect.FieldDescriptor
	for _, paramPart := range strings.Split(param, ".") {
		field := fieldByName(md, paramPart)
		if field == nil {
			return nil
		}
		current = field
	}
	return current
}

func fieldByName(md protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.JSONName() != name {
			continue
		}
		return field
	}
	return nil
}

func partsToParameter(tokens []Token) []string {
	params := []string{}
	for _, token := range tokens {
		if token.Type == TokenVariable {
			params = append(params, token.Value)
		}
	}
	return params
}

func partsToOpenAPIPath(tokens []Token) string {
	var b strings.Builder
	for _, token := range tokens {
		switch token.Type {
		case TokenSlash:
			b.WriteByte('/')
		case TokenEOF:
		case TokenLiteral:
			b.WriteString(token.Value)
		case TokenIdent:
			b.WriteString(token.Value)
		case TokenVariable:
			b.WriteByte('{')
			b.WriteString(token.Value)
			b.WriteByte('}')
		}
	}
	return b.String()
}

func schemaToMap(schema *jsonschema.Schema) map[string]interface{} {
	if schema == nil {
		return nil
	}
	return map[string]interface{}{
		"type":        schema.Type,
		"format":      schema.Format,
		"oneOf":       schema.OneOf,
		"ref":         schema.Ref,
		"title":       schema.Title,
		"description": schema.Description,
	}
}
