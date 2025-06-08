package googleapi

import (
	"fmt"
	"iter"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
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

func mergeOrAppendParameter(existingParams []*v3.Parameter, newParam *v3.Parameter) []*v3.Parameter {
	found := false
	for _, p := range existingParams {
		if p.Name != newParam.Name || p.In != newParam.In {
			continue
		}
		found = true
		if p.Description == "" && newParam.Description != "" {
			p.Description = newParam.Description
		}
		// If p.Required is nil (not set) and newParam.Required is set, then use newParam.Required.
		// This preserves an explicitly set false in p.Required.
		if p.Required == nil && newParam.Required != nil {
			p.Required = newParam.Required
		}
		if p.Schema == nil && newParam.Schema != nil {
			p.Schema = newParam.Schema
		} else if p.Schema != nil && newParam.Schema != nil {
			// Merge schema properties
			if p.Schema.Schema().Title == "" {
				p.Schema.Schema().Title = newParam.Schema.Schema().Title
			}
			if p.Schema.Schema().Description == "" {
				p.Schema.Schema().Description = newParam.Schema.Schema().Description
			}
			if len(p.Schema.Schema().Type) == 0 {
				p.Schema.Schema().Type = newParam.Schema.Schema().Type
			}
			if p.Schema.Schema().Format == "" {
				p.Schema.Schema().Format = newParam.Schema.Schema().Format
			}
			if len(p.Schema.Schema().Enum) == 0 {
				p.Schema.Schema().Enum = newParam.Schema.Schema().Enum
			}
			if p.Schema.Schema().Default == nil {
				p.Schema.Schema().Default = newParam.Schema.Schema().Default
			}
			if p.Schema.Schema().Items == nil {
				p.Schema.Schema().Items = newParam.Schema.Schema().Items
			}
		}
		// If p.Explode is nil (not set) and newParam.Explode is set, then use newParam.Explode.
		// This preserves an explicitly set false in p.Explode.
		if p.Explode == nil {
			p.Explode = newParam.Explode
		}
		// Assuming Deprecated, AllowEmptyValue, AllowReserved are bool (non-pointer) based on compiler errors
		// This means "empty/nil" is false. We update if current is false.
		if !p.Deprecated { // If p.Deprecated is false
			p.Deprecated = newParam.Deprecated // Set it from newParam
		}
		if !p.AllowEmptyValue { // If p.AllowEmptyValue is false
			p.AllowEmptyValue = newParam.AllowEmptyValue // Set it from newParam
		}
		if p.Style == "" {
			p.Style = newParam.Style
		}
		if !p.AllowReserved { // If p.AllowReserved is false
			p.AllowReserved = newParam.AllowReserved // Set it from newParam
		}
	}
	if !found {
		existingParams = append(existingParams, newParam)
	}
	return existingParams
}

// namedPathPattern is a regular expression to match named path patterns in the form {name=path/*/pattern}
var namedPathPattern = regexp.MustCompile("{(.+)=(.+)}")

func MakePathItems(opts options.Options, md protoreflect.MethodDescriptor) *orderedmap.Map[string, *v3.PathItem] {
	if opts.IgnoreGoogleapiHTTP {
		return nil
	}
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

	operationId := string(md.FullName())
	if opts.ShortOperationIds {
		operationId = string(service.Name()) + "_" + string(md.Name())
	}
	op := &v3.Operation{
		Summary:     string(md.Name()),
		OperationId: operationId,
		Description: util.FormatComments(fd.SourceLocations().ByDescriptor(md)),
	}

	if !opts.WithoutDefaultTags {
		tagName := string(service.FullName())
		if opts.ShortServiceTags {
			tagName = string(service.Name())
		}
		op.Tags = []string{tagName}
	}

	fieldNamesInPath := map[string]struct{}{}
	for _, param := range partsToParameter(tokens) {
		// Skip the name parameter if it's part of a glob pattern
		if strings.Contains(param, "=") {
			continue
		}
		field, jsonPath := resolveField(md.Input(), param)
		if field != nil {
			// This field is only top level, so we will filter out the param from
			// query/param or request body
			fieldNamesInPath[string(field.FullName())] = struct{}{}
			fieldNamesInPath[strings.Join(jsonPath, ".")] = struct{}{} // sometimes JSON field names are used
			loc := fd.SourceLocations().ByDescriptor(field)
			newParameter := &v3.Parameter{
				Name:        param,
				Required:    proto.Bool(true),
				In:          "path",
				Description: util.FormatComments(loc),
				Schema:      schema.FieldToSchema(opts, nil, field),
			}
			op.Parameters = mergeOrAppendParameter(op.Parameters, newParameter)
		} else {
			slog.Warn("path field not found", slog.String("param", param))
		}
	}

	// Add named path parameters from glob patterns
	for _, token := range tokens {
		if token.Type == TokenVariable && strings.Contains(token.Value, "=") {
			matches := namedPathPattern.FindStringSubmatch("{" + token.Value + "}")
			if len(matches) == 3 {
				// Store the original field name from the glob pattern to prevent it from appearing
				// in both the path parameters and request body/query parameters
				orignalName := matches[1]
				fieldNamesInPath[orignalName] = struct{}{}
				// Convert the path from the starred form to use named path parameters.
				starredPath := matches[2]
				parts := strings.Split(starredPath, "/")
				// The starred path is assumed to be in the form "things/*/otherthings/*".
				// We want to convert it to "things/{thingsId}/otherthings/{otherthingsId}".
				for i := 0; i < len(parts)-1; i += 2 {
					section := parts[i]
					namedPathParameter := util.Singular(section)
					// Add the parameter to the operation
					newParameter := &v3.Parameter{
						Name:        namedPathParameter,
						In:          "path",
						Required:    proto.Bool(true),
						Description: "The " + namedPathParameter + " id.",
						Schema:      base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}),
					}
					op.Parameters = mergeOrAppendParameter(op.Parameters, newParameter)
				}
			}
		}
	}

	switch rule.Body {
	case "":
		newQueryParams := flattenToParams(opts, md.Input(), "", fieldNamesInPath)
		for _, newQueryParam := range newQueryParams {
			op.Parameters = mergeOrAppendParameter(op.Parameters, newQueryParam)
		}
	case "*":
		if len(fieldNamesInPath) > 0 {
			_, s := schema.MessageToSchema(opts, md.Input())
			if s != nil && s.Properties != nil {
				for name := range fieldNamesInPath {
					s.Properties.Delete(name)
					// Also remove from required list to prevent duplicate required properties
					if s.Required != nil {
						s.Required = slices.DeleteFunc(s.Required, func(s string) bool {
							return s == name
						})
						// don't serialize []
						if len(s.Required) == 0 {
							s.Required = nil
						}
					}
				}
				if s.Properties.Len() > 0 {
					op.RequestBody = util.MethodToRequestBody(opts, md, base.CreateSchemaProxy(s), false)
				}
			}
		} else {
			inputName := string(md.Input().FullName())
			s := base.CreateSchemaProxyRef("#/components/schemas/" + util.FormatTypeRef(inputName))
			op.RequestBody = util.MethodToRequestBody(opts, md, s, false)
		}

	default:
		if field, _ := resolveField(md.Input(), rule.Body); field != nil {
			loc := fd.SourceLocations().ByDescriptor(field)
			bodySchema := schema.FieldToSchema(opts, nil, field)
			op.RequestBody = &v3.RequestBody{
				Description: util.FormatComments(loc),
				Content:     util.MakeMediaTypes(opts, bodySchema, false, false),
			}
		} else {
			slog.Warn("body field not found", slog.String("param", rule.Body))
		}
	}

	// Responses
	codeMap := orderedmap.New[string, *v3.Response]()
	mediaType := orderedmap.New[string, *v3.MediaType]()
	var outputSchema *base.SchemaProxy
	if rule.ResponseBody == "" {
		outputSchema = base.CreateSchemaProxyRef("#/components/schemas/" + util.FormatTypeRef(string(md.Output().FullName())))
	} else {
		if fd, _ := resolveField(md.Output(), rule.ResponseBody); fd != nil {
			outputSchema = schema.FieldToSchema(opts, nil, fd)
		}
	}

	mediaType.Set("application/json", &v3.MediaType{Schema: outputSchema})
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
			path := util.MakePath(opts, pair.Key())
			paths.Set(path, pair.Value())
		}
	}
	dedupeOperations(op.OperationId, paths.ValuesFromOldest())
	return paths
}

// dedupeOperations assigns unique operation ids to additional bindings.
// From the OpenAPI v3 spec: "The id MUST be unique among all operations described in the API."
// Since the same gRPC method name is used for operationId, the additional bindings will not be unique,
// so we append a number, starting at 2, when more than one path binds to the same method.
func dedupeOperations(id string, value iter.Seq[*v3.PathItem]) {
	num := 0
	for path := range value {
		for op := range path.GetOperations().ValuesFromOldest() {
			if op.OperationId == id {
				num++
				if num > 1 {
					op.OperationId = fmt.Sprintf("%s%d", id, num)
				}
			}
		}
	}
}

func resolveField(md protoreflect.MessageDescriptor, param string) (protoreflect.FieldDescriptor, []string) {
	jsonParts := []string{}
	current := md
	var fd protoreflect.FieldDescriptor
	for _, paramPart := range strings.Split(param, ".") {
		if field := fieldByName(current, paramPart); field == nil {
			return nil, nil
		} else {
			fd = field
			jsonParts = append(jsonParts, fd.JSONName())
			current = field.Message()
		}
	}
	return fd, jsonParts
}

func fieldByName(md protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	slog.Info("fieldByName", "name", md.FullName(), "name", name)
	fields := md.Fields()
	if field := fields.ByName(protoreflect.Name(name)); field != nil {
		return field
	}
	if field := fields.ByJSONName(name); field != nil {
		return field
	}
	return nil
}

func partsToParameter(tokens []Token) []string {
	params := []string{}
	for _, token := range tokens {
		if token.Type == TokenVariable {
			// Skip parameters that contain = as they are part of a glob pattern
			if strings.Contains(token.Value, "=") {
				continue
			}
			params = append(params, strings.SplitN(token.Value, "=", 2)[0])
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
			// Handle the name= prefix by extracting just the path portion
			if strings.Contains(token.Value, "=") {
				matches := namedPathPattern.FindStringSubmatch("{" + token.Value + "}")
				if len(matches) == 3 {
					// Add the "name=" "name" value to the list of covered parameters.
					// Convert the path from the starred form to use named path parameters.
					starredPath := matches[2]
					parts := strings.Split(starredPath, "/")
					// The starred path is assumed to be in the form "things/*/otherthings/*".
					// We want to convert it to "things/{thingsId}/otherthings/{otherthingsId}".
					for i := 0; i < len(parts)-1; i += 2 {
						section := parts[i]
						namedPathParameter := util.Singular(section)
						parts[i+1] = "{" + namedPathParameter + "}"
					}
					// Rewrite the path to use the path parameters.
					newPath := strings.Join(parts, "/")
					b.WriteString(newPath)
				} else {
					b.WriteByte('{')
					b.WriteString(token.Value)
					b.WriteByte('}')
				}
			} else {
				b.WriteByte('{')
				b.WriteString(token.Value)
				b.WriteByte('}')
			}
		}
	}
	return b.String()
}

func flattenToParams(opts options.Options, md protoreflect.MessageDescriptor, prefix string, seen map[string]struct{}) []*v3.Parameter {
	params := []*v3.Parameter{}
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		paramName := prefix + util.MakeFieldName(opts, field)
		// exclude fields already found in the path
		if _, ok := seen[string(field.FullName())]; ok {
			continue
		}
		if _, ok := seen[paramName]; ok {
			continue
		}
		seen[string(field.FullName())] = struct{}{}
		switch field.Kind() {
		case protoreflect.MessageKind:
			params = append(params, flattenToParams(opts, field.Message(), paramName+".", seen)...)
		default:
			parent := &base.Schema{}
			schema := schema.FieldToSchema(opts, base.CreateSchemaProxy(parent), field)
			var required *bool
			if len(parent.Required) > 0 {
				required = util.BoolPtr(true)
			}
			loc := field.ParentFile().SourceLocations().ByDescriptor(field)
			params = append(params, &v3.Parameter{
				Name:        paramName,
				In:          "query",
				Description: util.FormatComments(loc),
				Schema:      schema,
				Required:    required,
			})
		}
	}
	return params
}
