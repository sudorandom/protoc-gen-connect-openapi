package googleapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/schema"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func MakePathItems(opts options.Options, md protoreflect.MethodDescriptor) *orderedmap.Map[string, *v3.PathItem] {
	mdopts := md.Options()
	if !proto.HasExtension(mdopts, annotations.E_Http) {
		return nil
	}
	rule, ok := proto.GetExtension(mdopts, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return nil
	}
	return httpRuleToPathMap(opts, md, rule)
}

func httpRuleToPathMap(opts options.Options, md protoreflect.MethodDescriptor, rule *annotations.HttpRule) *orderedmap.Map[string, *v3.PathItem] {
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

	paths := orderedmap.New[string, *v3.PathItem]()
	pathItem := &v3.PathItem{}

	fd := md.ParentFile()
	service := md.Parent().(protoreflect.ServiceDescriptor)
	loc := fd.SourceLocations().ByDescriptor(md)
	op := &v3.Operation{
		Summary:     string(md.Name()),
		OperationId: string(md.FullName()),
		Tags:        []string{string(service.FullName())},
		Description: util.FormatComments(loc),
	}

	topLevelFieldNamesInPath := map[string]struct{}{}
	for _, param := range partsToParameter(tokens) {
		field := resolveField(md.Input(), param)
		if field != nil {
			parts := strings.Split(param, ".")
			// This field is only top level, so we will filter out the param from
			// query/param or request body
			if len(parts) == 1 {
				topLevelFieldNamesInPath[parts[0]] = struct{}{}
				topLevelFieldNamesInPath[field.JSONName()] = struct{}{} // sometimes JSON field names are used
			}
			loc := fd.SourceLocations().ByDescriptor(field)
			op.Parameters = append(op.Parameters, &v3.Parameter{
				Name:        param,
				Required:    proto.Bool(true),
				In:          "path",
				Description: util.FormatComments(loc),
				Schema:      schema.FieldToSchema(opts, nil, field),
			})
		}
	}

	switch rule.Body {
	case "":
		fields := md.Input().Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			// exclude fields already found in the path
			if _, ok := topLevelFieldNamesInPath[string(field.Name())]; ok {
				continue
			}
			if _, ok := topLevelFieldNamesInPath[field.JSONName()]; ok {
				continue
			}
			loc := fd.SourceLocations().ByDescriptor(md)
			desc := util.FormatComments(loc)
			op.Parameters = append(op.Parameters, &v3.Parameter{
				Name:        field.JSONName(),
				In:          "query",
				Description: desc,
				Schema:      schema.FieldToSchema(opts, nil, field),
			})
		}
	case "*":
		if len(topLevelFieldNamesInPath) > 0 {
			_, s := schema.MessageToSchema(opts, md.Input())
			for name := range topLevelFieldNamesInPath {
				s.Properties.Delete(name)
			}
			if s.Properties.Len() > 0 {
				op.RequestBody = util.MethodToRequestBody(opts, md, base.CreateSchemaProxy(s), false)
			}
		} else {
			inputName := string(md.Input().FullName())
			s := base.CreateSchemaProxyRef("#/components/schemas/" + util.FormatTypeRef(inputName))
			op.RequestBody = util.MethodToRequestBody(opts, md, s, false)
		}

	default:
		fields := md.Input().Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if field.JSONName() != rule.Body {
				continue
			}
			loc := fd.SourceLocations().ByDescriptor(md)
			op.RequestBody = &v3.RequestBody{
				Description: util.FormatComments(loc),
			}
		}
	}

	// Responses
	codeMap := orderedmap.New[string, *v3.Response]()

	id := util.FormatTypeRef(string(md.Output().FullName()))
	mediaType := orderedmap.New[string, *v3.MediaType]()
	mediaType.Set("application/json", &v3.MediaType{
		Schema: base.CreateSchemaProxyRef("#/components/schemas/" + id),
	})
	codeMap.Set("200", &v3.Response{
		Description: "Success",
		Content:     mediaType,
	})

	op.Responses = &v3.Responses{
		Codes: codeMap,
		Default: &v3.Response{
			Description: "Error",
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/connect.error"),
				false,
				false,
			),
		},
	}

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
	}
	paths.Set(partsToOpenAPIPath(tokens), pathItem)

	for _, binding := range rule.AdditionalBindings {
		pathMap := httpRuleToPathMap(opts, md, binding)
		for pair := pathMap.First(); pair != nil; pair = pair.Next() {
			paths.Set(pair.Key(), pair.Value())
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
		if field.JSONName() != name && string(field.Name()) != name {
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
		case TokenColon:
			b.WriteByte(':')
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
