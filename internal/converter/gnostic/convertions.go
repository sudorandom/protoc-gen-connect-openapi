package gnostic

import (
	"strconv"

	goa3 "github.com/google/gnostic/openapiv3"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"go.yaml.in/yaml/v4"
)

func toServers(servers []*goa3.Server) []*v3.Server {
	if len(servers) == 0 {
		return nil
	}
	result := make([]*v3.Server, len(servers))
	for i, server := range servers {
		result[i] = toServer(server)
	}
	return result
}

func toServer(server *goa3.Server) *v3.Server {
	if server == nil {
		return nil
	}
	return &v3.Server{
		URL:         server.Url,
		Description: server.Description,
		Variables:   toVariables(server.Variables),
	}
}

func toVariables(variables *goa3.ServerVariables) *orderedmap.Map[string, *v3.ServerVariable] {
	if variables == nil || len(variables.AdditionalProperties) == 0 {
		return nil
	}
	vars := orderedmap.New[string, *v3.ServerVariable]()
	for _, prop := range variables.AdditionalProperties {
		vars.Store(prop.Name, &v3.ServerVariable{
			Enum:        prop.Value.Enum,
			Default:     prop.Value.Default,
			Description: prop.Value.Description,
		})
	}
	return vars
}

func toSecurityRequirements(securityReq []*goa3.SecurityRequirement) []*base.SecurityRequirement {
	result := make([]*base.SecurityRequirement, len(securityReq))
	for i, req := range securityReq {
		reqs := orderedmap.New[string, []string]()
		for _, prop := range req.AdditionalProperties {
			reqs.Set(prop.Name, prop.Value.Value)
		}

		result[i] = &base.SecurityRequirement{
			Requirements:             reqs,
			ContainsEmptyRequirement: len(req.AdditionalProperties) == 0,
		}
	}
	return result
}

func appendComponents(opts options.Options, spec *v3.Document, c *goa3.Components) {
	if c == nil {
		return
	}
	util.AppendComponents(spec, &v3.Components{
		Schemas:         toSchemaOrReferenceMap(opts, c.Schemas.GetAdditionalProperties()),
		SecuritySchemes: toSecuritySchemes(c.SecuritySchemes),
		Responses:       toResponsesMap(opts, c.Responses),
		Parameters:      toParametersMap(opts, c.Parameters),
		Examples:        toExamples(c.Examples),
		RequestBodies:   toRequestBodiesMap(opts, c.RequestBodies),
		Headers:         toHeaders(opts, c.Headers),
		Links:           toLinks(c.Links),
		Callbacks:       toCallbacks(opts, c.Callbacks),
		Extensions:      toExtensions(c.SpecificationExtension),
	})
}

func toParametersMap(opts options.Options, params *goa3.ParametersOrReferences) *orderedmap.Map[string, *v3.Parameter] {
	m := orderedmap.New[string, *v3.Parameter]()
	for _, item := range params.GetAdditionalProperties() {
		m.Set(item.Name, toParameter(opts, item.GetValue()))
	}
	return m
}

func toRequestBodiesMap(opts options.Options, bodies *goa3.RequestBodiesOrReferences) *orderedmap.Map[string, *v3.RequestBody] {
	m := orderedmap.New[string, *v3.RequestBody]()
	for _, item := range bodies.GetAdditionalProperties() {
		m.Set(item.Name, toRequestBody(opts, item.GetValue().GetRequestBody()))
	}
	return m
}

func toRequestBody(opts options.Options, rbody *goa3.RequestBody) *v3.RequestBody {
	return &v3.RequestBody{
		Description: rbody.Description,
		Content:     toMediaTypes(opts, rbody.GetContent()),
		Required:    &rbody.Required,
		Extensions:  toExtensions(rbody.SpecificationExtension),
	}
}

func toResponsesMap(opts options.Options, resps *goa3.ResponsesOrReferences) *orderedmap.Map[string, *v3.Response] {
	if resps == nil {
		return nil
	}
	m := orderedmap.New[string, *v3.Response]()
	for _, resp := range resps.GetAdditionalProperties() {
		m.Set(resp.Name, toResponse(opts, resp.Value))
	}
	return m
}

//gocyclo:ignore
func toSecuritySchemes(s *goa3.SecuritySchemesOrReferences) *orderedmap.Map[string, *v3.SecurityScheme] {
	if s == nil {
		return nil
	}

	secSchemas := orderedmap.New[string, *v3.SecurityScheme]()
	for _, addProp := range s.AdditionalProperties {
		secScheme := addProp.Value.GetSecurityScheme()
		if secScheme != nil {
			scheme := &v3.SecurityScheme{
				Name:             secScheme.Name,
				Description:      secScheme.Description,
				Type:             secScheme.Type,
				Scheme:           secScheme.Scheme,
				BearerFormat:     secScheme.BearerFormat,
				In:               secScheme.In,
				OpenIdConnectUrl: secScheme.OpenIdConnectUrl,
			}
			if secScheme.Flows != nil {
				flows := &v3.OAuthFlows{
					Extensions: toExtensions(secScheme.Flows.SpecificationExtension),
				}
				if secScheme.Flows.Implicit != nil {
					scopes := orderedmap.New[string, string]()
					for _, scope := range secScheme.Flows.Implicit.Scopes.AdditionalProperties {
						scopes.Set(scope.Name, scope.Value)
					}
					flows.Implicit = &v3.OAuthFlow{
						AuthorizationUrl: secScheme.Flows.Implicit.AuthorizationUrl,
						TokenUrl:         secScheme.Flows.Implicit.TokenUrl,
						RefreshUrl:       secScheme.Flows.Implicit.RefreshUrl,
						Scopes:           scopes,
						Extensions:       toExtensions(secScheme.Flows.Implicit.SpecificationExtension),
					}
				}
				if secScheme.Flows.Password != nil {
					scopes := orderedmap.New[string, string]()
					for _, scope := range secScheme.Flows.Password.Scopes.AdditionalProperties {
						scopes.Set(scope.Name, scope.Value)
					}
					flows.Password = &v3.OAuthFlow{
						TokenUrl:   secScheme.Flows.Password.TokenUrl,
						RefreshUrl: secScheme.Flows.Password.RefreshUrl,
						Scopes:     scopes,
						Extensions: toExtensions(secScheme.Flows.Password.SpecificationExtension),
					}
				}
				if secScheme.Flows.ClientCredentials != nil {
					scopes := orderedmap.New[string, string]()
					for _, scope := range secScheme.Flows.ClientCredentials.Scopes.AdditionalProperties {
						scopes.Set(scope.Name, scope.Value)
					}
					flows.ClientCredentials = &v3.OAuthFlow{
						TokenUrl:   secScheme.Flows.ClientCredentials.TokenUrl,
						RefreshUrl: secScheme.Flows.ClientCredentials.RefreshUrl,
						Scopes:     scopes,
						Extensions: toExtensions(secScheme.Flows.ClientCredentials.SpecificationExtension),
					}
				}
				if secScheme.Flows.AuthorizationCode != nil {
					scopes := orderedmap.New[string, string]()
					for _, scope := range secScheme.Flows.AuthorizationCode.Scopes.AdditionalProperties {
						scopes.Set(scope.Name, scope.Value)
					}
					flows.AuthorizationCode = &v3.OAuthFlow{
						AuthorizationUrl: secScheme.Flows.AuthorizationCode.AuthorizationUrl,
						TokenUrl:         secScheme.Flows.AuthorizationCode.TokenUrl,
						RefreshUrl:       secScheme.Flows.AuthorizationCode.RefreshUrl,
						Scopes:           scopes,
						Extensions:       toExtensions(secScheme.Flows.AuthorizationCode.SpecificationExtension),
					}
				}
				scheme.Flows = flows
			}
			secSchemas.Set(addProp.Name, scheme)
		}
	}

	return secSchemas
}

func toExternalDocs(externalDocs *goa3.ExternalDocs) *base.ExternalDoc {
	if externalDocs == nil {
		return nil
	}

	return &base.ExternalDoc{
		Description: externalDocs.Description,
		URL:         externalDocs.Url,
		Extensions:  toExtensions(externalDocs.SpecificationExtension),
	}
}

func toTags(tags []*goa3.Tag) []*base.Tag {
	if len(tags) == 0 {
		return nil
	}

	result := make([]*base.Tag, len(tags))
	for i, tag := range tags {
		var extDoc *base.ExternalDoc
		if tag.ExternalDocs != nil {
			extDoc = &base.ExternalDoc{
				Description: tag.ExternalDocs.Description,
				URL:         tag.ExternalDocs.Url,
				Extensions:  toExtensions(tag.ExternalDocs.SpecificationExtension),
			}
		}
		result[i] = &base.Tag{
			Name:         tag.Name,
			Description:  tag.Description,
			ExternalDocs: extDoc,
			Extensions:   toExtensions(tag.SpecificationExtension),
		}
	}
	return result
}

func toSchemaOrReferences(opts options.Options, items []*goa3.SchemaOrReference) []*base.SchemaProxy {
	result := make([]*base.SchemaProxy, len(items))
	for i, s := range items {
		result[i] = toSchemaOrReference(opts, s)
	}
	return result
}

func toSchemaOrReference(opts options.Options, s *goa3.SchemaOrReference) *base.SchemaProxy {
	if s == nil {
		return nil
	}
	if ref := s.GetReference(); ref != nil {
		return base.CreateSchemaProxyRef(util.ResolveSchemaRef(ref.XRef))
	} else if schema := s.GetSchema(); schema != nil {
		return base.CreateSchemaProxy(toSchema(opts, schema))
	}
	return nil
}

func toSchemaOrReferenceMap(opts options.Options, items []*goa3.NamedSchemaOrReference) *orderedmap.Map[string, *base.SchemaProxy] {
	m := orderedmap.New[string, *base.SchemaProxy]()
	for _, item := range items {
		m.Set(item.Name, toSchemaOrReference(opts, item.Value))
	}
	return m
}

func toSchema(opts options.Options, s *goa3.Schema) *base.Schema {
	if s == nil {
		return nil
	}
	return schemaWithAnnotations(opts, &base.Schema{}, s)
}

func toDefault(dt *goa3.DefaultType) *yaml.Node {
	if dt == nil {
		return nil
	}
	switch dt.GetOneof().(type) {
	case *goa3.DefaultType_Number:
		return utils.CreateStringNode(strconv.FormatFloat(dt.GetNumber(), 'f', -1, 64))
	case *goa3.DefaultType_String_:
		return utils.CreateStringNode(dt.GetString_())
	case *goa3.DefaultType_Boolean:
		if dt.GetBoolean() {
			return utils.CreateStringNode("true")
		}
		return utils.CreateStringNode("false")
	default:
		return nil
	}
}

func toAdditionalPropertiesItem(opts options.Options, item *goa3.AdditionalPropertiesItem) *base.DynamicValue[*base.SchemaProxy, bool] {
	switch v := item.Oneof.(type) {
	case *goa3.AdditionalPropertiesItem_SchemaOrReference:
		return &base.DynamicValue[*base.SchemaProxy, bool]{A: toSchemaOrReference(opts, v.SchemaOrReference)}
	case *goa3.AdditionalPropertiesItem_Boolean:
		return &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: v.Boolean}
	}
	return nil
}

func toExtensions(items []*goa3.NamedAny) *orderedmap.Map[string, *yaml.Node] {
	if items == nil {
		return nil
	}
	extensions := orderedmap.New[string, *yaml.Node]()
	for _, namedAny := range items {
		extensions.Set(namedAny.Name, util.ConvertNodeV3toV4(namedAny.Value.ToRawInfo()))
	}
	return extensions
}

func toEncodings(opts options.Options, enc *goa3.Encodings) *orderedmap.Map[string, *v3.Encoding] {
	if enc == nil {
		return nil
	}
	encodings := orderedmap.New[string, *v3.Encoding]()
	for _, encoding := range enc.GetAdditionalProperties() {
		encodings.Set(encoding.Name, &v3.Encoding{
			ContentType:   encoding.Value.ContentType,
			Headers:       toHeaders(opts, encoding.Value.Headers),
			Style:         encoding.Value.Style,
			Explode:       &encoding.Value.Explode,
			AllowReserved: encoding.Value.AllowReserved,
		})
	}
	return encodings
}

func toExamples(exes *goa3.ExamplesOrReferences) *orderedmap.Map[string, *base.Example] {
	if exes == nil {
		return nil
	}
	examples := orderedmap.New[string, *base.Example]()
	for _, item := range exes.GetAdditionalProperties() {
		if example := item.GetValue().GetExample(); example != nil {
			examples.Set(item.Name, &base.Example{
				Summary:       example.Summary,
				Description:   example.Description,
				Value:         util.ConvertNodeV3toV4(example.Value.ToRawInfo()),
				ExternalValue: example.ExternalValue,
				Extensions:    toExtensions(example.SpecificationExtension),
			})
		}
		if ref := item.GetValue().GetReference(); ref != nil {
			example := base.CreateExampleRef(ref.XRef)
			example.Summary = ref.Summary
			example.Description = ref.Description
			examples.Set(item.Name, example)
		}
	}
	return examples
}

func toMediaTypes(opts options.Options, items *goa3.MediaTypes) *orderedmap.Map[string, *v3.MediaType] {
	if items == nil {
		return nil
	}
	content := orderedmap.New[string, *v3.MediaType]()
	for _, item := range items.GetAdditionalProperties() {
		mt := &v3.MediaType{
			Schema:     toSchemaOrReference(opts, item.Value.Schema),
			Examples:   toExamples(item.Value.GetExamples()),
			Encoding:   toEncodings(opts, item.Value.GetEncoding()),
			Extensions: toExtensions(item.Value.GetSpecificationExtension()),
		}
		if val := item.GetValue().Example; val != nil {
			mt.Example = util.ConvertNodeV3toV4(val.ToRawInfo())
		}
		content.Set(item.Name, mt)
	}
	return content
}

func toHeaders(opts options.Options, v *goa3.HeadersOrReferences) *orderedmap.Map[string, *v3.Header] {
	if v == nil {
		return nil
	}
	headers := orderedmap.New[string, *v3.Header]()
	for _, headerVal := range v.GetAdditionalProperties() {
		if ref := headerVal.Value.GetReference(); ref != nil {
			header := v3.CreateHeaderRef(ref.XRef)
			header.Description = ref.Description
			headers.Set(headerVal.Name, header)
		} else if header := headerVal.Value.GetHeader(); header != nil {
			var exampleRawInfo *yaml.Node
			if header.Example != nil {
				exampleRawInfo = util.ConvertNodeV3toV4(header.Example.ToRawInfo())
			}
			headers.Set(headerVal.Name, &v3.Header{
				Description:     header.Description,
				Required:        header.Required,
				Deprecated:      header.Deprecated,
				AllowEmptyValue: header.AllowEmptyValue,
				Style:           header.Style,
				Explode:         header.Explode,
				AllowReserved:   header.AllowReserved,
				Schema:          toSchemaOrReference(opts, header.Schema),
				Example:         exampleRawInfo,
				Examples:        toExamples(header.GetExamples()),
				Content:         toMediaTypes(opts, header.Content),
				Extensions:      toExtensions(header.GetSpecificationExtension()),
			})
		}
	}
	return headers
}

func toCodes(opts options.Options, responses []*goa3.NamedResponseOrReference) *orderedmap.Map[string, *v3.Response] {
	resps := orderedmap.New[string, *v3.Response]()
	for _, resp := range responses {
		resps.Set(resp.Name, toResponse(opts, resp.GetValue()))
	}
	return resps
}

func toResponses(opts options.Options, responses *goa3.Responses) *v3.Responses {
	if responses == nil {
		return nil
	}
	return &v3.Responses{
		Codes:      toCodes(opts, responses.GetResponseOrReference()),
		Default:    toResponse(opts, responses.GetDefault()),
		Extensions: toExtensions(responses.GetSpecificationExtension()),
	}
}

func toResponse(opts options.Options, r *goa3.ResponseOrReference) *v3.Response {
	if r == nil {
		return nil
	}
	if v := r.GetReference(); v != nil {
		response := v3.CreateResponseRef(v.XRef)
		response.Description = v.Description
		return response
	}
	if v := r.GetResponse(); v != nil {
		return &v3.Response{
			Description: v.Description,
			Headers:     toHeaders(opts, v.Headers),
			Content:     toMediaTypes(opts, v.Content),
			Links:       toLinks(v.Links),
			Extensions:  toExtensions(v.SpecificationExtension),
		}
	}

	return nil
}

func toLinks(ls *goa3.LinksOrReferences) *orderedmap.Map[string, *v3.Link] {
	if ls == nil {
		return nil
	}
	links := orderedmap.New[string, *v3.Link]()
	for _, item := range ls.AdditionalProperties {
		link := item.Value.GetLink()
		params := orderedmap.New[string, string]()
		for _, param := range link.Parameters.GetExpression().GetAdditionalProperties() {
			params.Set(param.Name, param.Value.Yaml)
		}

		links.Set(item.Name, &v3.Link{
			OperationRef: link.OperationRef,
			OperationId:  link.OperationId,
			Parameters:   params,
			RequestBody:  link.RequestBody.String(),
			Description:  link.Description,
			Server:       toServer(link.Server),
			Extensions:   toExtensions(link.SpecificationExtension),
		})
	}
	return links
}

func toCallbacks(opts options.Options, cbs *goa3.CallbacksOrReferences) *orderedmap.Map[string, *v3.Callback] {
	if cbs == nil {
		return nil
	}
	callbacks := orderedmap.New[string, *v3.Callback]()
	for _, item := range cbs.GetAdditionalProperties() {
		callback := item.Value.GetCallback()
		expressions := orderedmap.New[string, *v3.PathItem]()
		for _, item := range callback.GetPath() {
			expr := item.Value
			expressions.Set(item.Name, &v3.PathItem{
				Description: expr.Description,
				Summary:     expr.Summary,
				Get:         toOperation(opts, expr.Get),
				Put:         toOperation(opts, expr.Put),
				Post:        toOperation(opts, expr.Post),
				Delete:      toOperation(opts, expr.Delete),
				Options:     toOperation(opts, expr.Options),
				Head:        toOperation(opts, expr.Head),
				Patch:       toOperation(opts, expr.Patch),
				Trace:       toOperation(opts, expr.Trace),
				Servers:     toServers(expr.Servers),
				Parameters:  toParameters(opts, expr.Parameters),
				Extensions:  toExtensions(expr.SpecificationExtension),
			})
		}
		callbacks.Set(item.Name, &v3.Callback{
			Expression: expressions,
			Extensions: toExtensions(callback.GetSpecificationExtension()),
		})
	}
	return callbacks
}

func toOperation(opts options.Options, op *goa3.Operation) *v3.Operation {
	if op == nil {
		return nil
	}
	return &v3.Operation{
		Tags:         op.Tags,
		Summary:      op.Summary,
		Description:  op.Description,
		ExternalDocs: toExternalDocs(op.ExternalDocs),
		OperationId:  op.OperationId,
		Parameters:   toParameters(opts, op.Parameters),
		RequestBody:  nil,
		Responses:    toResponses(opts, op.GetResponses()),
		Callbacks:    toCallbacks(opts, op.Callbacks),
		Deprecated:   &op.Deprecated,
		Security:     toSecurityRequirements(op.Security),
		Servers:      toServers(op.Servers),
		Extensions:   toExtensions(op.SpecificationExtension),
	}
}

func toParameters(opts options.Options, params []*goa3.ParameterOrReference) []*v3.Parameter {
	if params == nil {
		return nil
	}
	parameters := make([]*v3.Parameter, len(params))
	for i, param := range params {
		parameters[i] = toParameter(opts, param)
	}
	return parameters
}

func toParameter(opts options.Options, paramOrRef *goa3.ParameterOrReference) *v3.Parameter {
	if paramOrRef == nil || paramOrRef.GetParameter() == nil {
		return nil
	}
	param := paramOrRef.GetParameter()

	var example *yaml.Node
	if param.Example != nil {
		example = util.ConvertNodeV3toV4(param.Example.ToRawInfo())
	}

	return &v3.Parameter{
		Name:            param.GetName(),
		In:              param.In,
		Description:     param.Description,
		Required:        &param.Required,
		Deprecated:      param.Deprecated,
		AllowEmptyValue: param.AllowEmptyValue,
		Style:           param.Style,
		Explode:         &param.Explode,
		AllowReserved:   param.AllowReserved,
		Schema:          toSchemaOrReference(opts, param.GetSchema()),
		Example:         example,
		Content:         toMediaTypes(opts, param.GetContent()),
		Extensions:      toExtensions(param.GetSpecificationExtension()),
	}
}
