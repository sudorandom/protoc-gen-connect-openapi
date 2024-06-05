package gnostic

import (
	"strconv"

	goa3 "github.com/google/gnostic/openapiv3"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"gopkg.in/yaml.v3"
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

func toComponents(c *goa3.Components) *v3.Components {
	if c == nil {
		return nil
	}
	return &v3.Components{
		Schemas:         toSchemaOrReferenceMap(c.Schemas.GetAdditionalProperties()),
		SecuritySchemes: toSecuritySchemes(c.SecuritySchemes),
		Responses:       toResponsesMap(c.Responses),
		Parameters:      toParametersMap(c.Parameters),
		Examples:        toExamples(c.Examples),
		RequestBodies:   toRequestBodiesMap(c.RequestBodies),
		Headers:         toHeaders(c.Headers),
		Links:           toLinks(c.Links),
		Callbacks:       toCallbacks(c.Callbacks),
		Extensions:      toExtensions(c.SpecificationExtension),
	}
}

func toParametersMap(params *goa3.ParametersOrReferences) *orderedmap.Map[string, *v3.Parameter] {
	m := orderedmap.New[string, *v3.Parameter]()
	for _, item := range params.GetAdditionalProperties() {
		m.Set(item.Name, toParameter(item.GetValue()))
	}
	return m
}

func toRequestBodiesMap(bodies *goa3.RequestBodiesOrReferences) *orderedmap.Map[string, *v3.RequestBody] {
	m := orderedmap.New[string, *v3.RequestBody]()
	for _, item := range bodies.GetAdditionalProperties() {
		m.Set(item.Name, toRequestBody(item.GetValue().GetRequestBody()))
	}
	return m
}

func toRequestBody(rbody *goa3.RequestBody) *v3.RequestBody {
	return &v3.RequestBody{
		Description: rbody.Description,
		Content:     toMediaTypes(rbody.GetContent()),
		Required:    &rbody.Required,
		Extensions:  toExtensions(rbody.SpecificationExtension),
	}
}

func toResponsesMap(resps *goa3.ResponsesOrReferences) *orderedmap.Map[string, *v3.Response] {
	if resps == nil {
		return nil
	}
	m := orderedmap.New[string, *v3.Response]()
	for _, resp := range resps.GetAdditionalProperties() {
		m.Set(resp.Name, toResponse(resp.Value.GetResponse()))
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
				Type: secScheme.Type,
			}
			switch secScheme.Type {
			case "http":
				scheme.Scheme = secScheme.Scheme
			case "apiKey":
				scheme.Name = secScheme.Name
				scheme.In = secScheme.In
			case "openIdConnect":
				scheme.OpenIdConnectUrl = secScheme.OpenIdConnectUrl
			case "oauth2":
				flows := &v3.OAuthFlows{}
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
					}
				}
				scheme.Flows = flows
			default:
				continue
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
				Description: tag.Description,
				URL:         tag.ExternalDocs.Url,
			}
		}
		result[i] = &base.Tag{
			Name:         tag.Name,
			Description:  tag.Description,
			ExternalDocs: extDoc,
		}
	}
	return result
}

func toSchemaOrReferences(items []*goa3.SchemaOrReference) []*base.SchemaProxy {
	result := make([]*base.SchemaProxy, len(items))
	for i, s := range items {
		result[i] = toSchemaOrReference(s)
	}
	return result
}

func toSchemaOrReference(s *goa3.SchemaOrReference) *base.SchemaProxy {
	if s == nil {
		return nil
	}
	if ref := s.GetReference(); ref != nil {
		return base.CreateSchemaProxyRef(ref.XRef)
	} else if schema := s.GetSchema(); schema != nil {
		return base.CreateSchemaProxy(toSchema(schema))
	}
	return nil
}

func toSchemaOrReferenceMap(items []*goa3.NamedSchemaOrReference) *orderedmap.Map[string, *base.SchemaProxy] {
	m := orderedmap.New[string, *base.SchemaProxy]()
	for _, item := range items {
		m.Set(item.Name, toSchemaOrReference(item.Value))
	}
	return m
}

func toSchema(s *goa3.Schema) *base.Schema {
	if s == nil {
		return nil
	}
	return schemaWithAnnotations(&base.Schema{}, s)
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

func toAdditionalPropertiesItem(item *goa3.AdditionalPropertiesItem) *base.DynamicValue[*base.SchemaProxy, bool] {
	switch v := item.Oneof.(type) {
	case *goa3.AdditionalPropertiesItem_SchemaOrReference:
		return &base.DynamicValue[*base.SchemaProxy, bool]{A: toSchemaOrReference(v.SchemaOrReference)}
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
		extensions.Set(namedAny.Name, namedAny.ToRawInfo())
	}
	return extensions
}

func toEncodings(enc *goa3.Encodings) *orderedmap.Map[string, *v3.Encoding] {
	if enc == nil {
		return nil
	}
	encodings := orderedmap.New[string, *v3.Encoding]()
	for _, encoding := range enc.GetAdditionalProperties() {
		encodings.Set(encoding.Name, &v3.Encoding{
			ContentType:   encoding.Value.ContentType,
			Headers:       toHeaders(encoding.Value.Headers),
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
		example := item.GetValue().GetExample()
		examples.Set(item.Name, &base.Example{
			Summary:       example.Summary,
			Description:   example.Description,
			Value:         example.Value.ToRawInfo(),
			ExternalValue: example.ExternalValue,
			Extensions:    toExtensions(example.SpecificationExtension),
		})
	}
	return examples
}

func toMediaTypes(items *goa3.MediaTypes) *orderedmap.Map[string, *v3.MediaType] {
	if items == nil {
		return nil
	}
	content := orderedmap.New[string, *v3.MediaType]()
	for _, item := range items.GetAdditionalProperties() {
		content.Set(item.Name, &v3.MediaType{
			Schema:     toSchemaOrReference(item.Value.Schema),
			Example:    item.Value.Example.ToRawInfo(),
			Examples:   toExamples(item.Value.GetExamples()),
			Encoding:   toEncodings(item.Value.GetEncoding()),
			Extensions: toExtensions(item.Value.GetSpecificationExtension()),
		})
	}
	return content
}

func toHeaders(v *goa3.HeadersOrReferences) *orderedmap.Map[string, *v3.Header] {
	if v == nil {
		return nil
	}
	headers := orderedmap.New[string, *v3.Header]()
	for _, headerVal := range v.GetAdditionalProperties() {
		header := headerVal.Value.GetHeader()
		headers.Set(headerVal.Name, &v3.Header{
			Description:     header.Description,
			Required:        header.Required,
			Deprecated:      header.Deprecated,
			AllowEmptyValue: header.AllowEmptyValue,
			Style:           header.Style,
			Explode:         header.Explode,
			AllowReserved:   header.AllowReserved,
			Schema:          toSchemaOrReference(header.Schema),
			Example:         header.Example.ToRawInfo(),
			Examples:        toExamples(header.GetExamples()),
			Content:         toMediaTypes(header.Content),
			Extensions:      toExtensions(header.GetSpecificationExtension()),
		})
	}
	return headers
}

func toCodes(responses []*goa3.NamedResponseOrReference) *orderedmap.Map[string, *v3.Response] {
	resps := orderedmap.New[string, *v3.Response]()
	for _, resp := range responses {
		resps.Set(resp.Name, toResponse(resp.Value.GetResponse()))
	}
	return resps
}

func toResponses(responses *goa3.Responses) *v3.Responses {
	if responses == nil {
		return nil
	}
	return &v3.Responses{
		Codes:      toCodes(responses.GetResponseOrReference()),
		Default:    toResponse(responses.GetDefault().GetResponse()),
		Extensions: toExtensions(responses.GetSpecificationExtension()),
	}
}

func toResponse(r *goa3.Response) *v3.Response {
	if r == nil {
		return nil
	}
	return &v3.Response{
		Description: r.Description,
		Headers:     toHeaders(r.Headers),
		Content:     toMediaTypes(r.Content),
		Links:       toLinks(r.Links),
		Extensions:  toExtensions(r.SpecificationExtension),
	}
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

func toCallbacks(cbs *goa3.CallbacksOrReferences) *orderedmap.Map[string, *v3.Callback] {
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
				Get:         toOperation(expr.Get),
				Put:         toOperation(expr.Put),
				Post:        toOperation(expr.Post),
				Delete:      toOperation(expr.Delete),
				Options:     toOperation(expr.Options),
				Head:        toOperation(expr.Head),
				Patch:       toOperation(expr.Patch),
				Trace:       toOperation(expr.Trace),
				Servers:     toServers(expr.Servers),
				Parameters:  toParameters(expr.Parameters),
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

func toOperation(op *goa3.Operation) *v3.Operation {
	if op == nil {
		return nil
	}
	return &v3.Operation{
		Tags:         op.Tags,
		Summary:      op.Summary,
		Description:  op.Description,
		ExternalDocs: toExternalDocs(op.ExternalDocs),
		OperationId:  op.OperationId,
		Parameters:   toParameters(op.Parameters),
		RequestBody:  nil,
		Responses:    toResponses(op.GetResponses()),
		Callbacks:    toCallbacks(op.Callbacks),
		Deprecated:   &op.Deprecated,
		Security:     toSecurityRequirements(op.Security),
		Servers:      toServers(op.Servers),
		Extensions:   toExtensions(op.SpecificationExtension),
	}
}

func toParameters(params []*goa3.ParameterOrReference) []*v3.Parameter {
	if params == nil {
		return nil
	}
	parameters := make([]*v3.Parameter, len(params))
	for i, param := range params {
		parameters[i] = toParameter(param)
	}
	return parameters
}

func toParameter(paramOrRef *goa3.ParameterOrReference) *v3.Parameter {
	if paramOrRef == nil || paramOrRef.GetParameter() == nil {
		return nil
	}
	param := paramOrRef.GetParameter()
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
		Schema:          toSchemaOrReference(param.GetSchema()),
		Example:         param.Example.ToRawInfo(),
		Content:         toMediaTypes(param.GetContent()),
		Extensions:      toExtensions(param.GetSpecificationExtension()),
	}
}
