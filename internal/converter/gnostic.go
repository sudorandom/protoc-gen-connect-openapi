package converter

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func pathItemWithMethodAnnotations(item openapi31.PathItem, md protoreflect.MethodDescriptor) openapi31.PathItem {
	if !proto.HasExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type()) {
		return item
	}

	ext := proto.GetExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Operation)
	if !ok {
		return item
	}
	for _, oper := range getAllOperations(item) {
		if opts.Deprecated {
			t := true
			oper.Deprecated = &t
		}

		security := []map[string][]string{}
		for _, sec := range opts.Security {
			m := map[string][]string{}
			for _, v := range sec.AdditionalProperties {
				m[v.Name] = v.Value.Value
			}
			security = append(security, m)
		}
		oper.Security = security
		if opts.Summary != "" {
			oper.Summary = &opts.Summary
		}
		if opts.Description != "" {
			oper.Description = &opts.Description
		}
	}
	return item
}

func getAllOperations(item openapi31.PathItem) []*openapi31.Operation {
	operations := []*openapi31.Operation{}
	if item.Get != nil {
		operations = append(operations, item.Get)
	}
	if item.Post != nil {
		operations = append(operations, item.Post)
	}
	if item.Put != nil {
		operations = append(operations, item.Put)
	}
	if item.Delete != nil {
		operations = append(operations, item.Delete)
	}
	if item.Head != nil {
		operations = append(operations, item.Head)
	}
	if item.Patch != nil {
		operations = append(operations, item.Patch)
	}
	if item.Options != nil {
		operations = append(operations, item.Options)
	}
	if item.Trace != nil {
		operations = append(operations, item.Trace)
	}
	return operations
}

func specWithFileAnnotations(spec *openapi31.Spec, fd protoreflect.FileDescriptor) *openapi31.Spec {
	if !proto.HasExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type()) {
		return spec
	}

	ext := proto.GetExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Document)
	if !ok {
		return spec
	}
	if opts.Openapi != "" {
		spec.Openapi = opts.Openapi
	}
	addInfoToSpec(spec, opts.Info)
	addServersToSpec(spec, opts.Servers)
	addSecurityToSpec(spec, opts.Security)
	addComponentsToSpec(spec, opts.Components)
	addTagsToSpec(spec, opts.Tags)
	addExternalDocsToSpec(spec, opts.ExternalDocs)
	return spec
}

func addInfoToSpec(spec *openapi31.Spec, info *goa3.Info) {
	if info == nil {
		return
	}
	spec.Info.Title = info.Title
	spec.Info.Summary = &info.Summary
	spec.Info.Description = &info.Description
	spec.Info.TermsOfService = &info.TermsOfService
	if info.Contact != nil {
		spec.Info.Contact = &openapi31.Contact{
			Name:  &info.Contact.Name,
			URL:   &info.Contact.Url,
			Email: &info.Contact.Email,
		}
	}
	if info.License != nil {
		spec.Info.License = &openapi31.License{
			Name: info.License.Name,
			URL:  &info.License.Url,
		}
	}
	spec.Info.Version = info.Version
}

func addServersToSpec(spec *openapi31.Spec, servers []*goa3.Server) {
	for _, server := range servers {
		spec.Servers = append(spec.Servers, openapi31.Server{
			URL:         server.Url,
			Description: &server.Description,
			Variables:   convertGnosticVariables(server.Variables),
		})
	}
}

func convertGnosticVariables(variables *goa3.ServerVariables) map[string]openapi31.ServerVariable {
	if variables == nil || len(variables.AdditionalProperties) == 0 {
		return nil
	}

	vars := make(map[string]openapi31.ServerVariable, len(variables.AdditionalProperties))
	for _, prop := range variables.AdditionalProperties {
		vars[prop.Name] = openapi31.ServerVariable{
			Enum:        prop.Value.Enum,
			Default:     prop.Value.Default,
			Description: &prop.Value.Description,
		}
	}
	return vars
}

func addSecurityToSpec(spec *openapi31.Spec, securityReq []*goa3.SecurityRequirement) {
	for _, req := range securityReq {
		item := map[string][]string{}
		for _, prop := range req.AdditionalProperties {
			item[prop.Name] = prop.Value.Value
		}

		spec.Security = append(spec.Security, item)
	}
}

func addComponentsToSpec(spec *openapi31.Spec, components *goa3.Components) {
	if components.SecuritySchemes != nil {
		secSchemas := map[string]openapi31.SecuritySchemeOrReference{}
		for _, addProp := range components.SecuritySchemes.AdditionalProperties {
			item := openapi31.SecuritySchemeOrReference{}
			ref := addProp.Value.GetReference()
			if ref != nil {
				item.Reference = &openapi31.Reference{
					Ref:         ref.XRef,
					Summary:     &ref.Summary,
					Description: &ref.Description,
				}
			}
			secScheme := addProp.Value.GetSecurityScheme()
			if secScheme != nil {
				scheme := &openapi31.SecurityScheme{}
				switch secScheme.Type {
				case "http":
					scheme.HTTP = &openapi31.SecuritySchemeHTTP{
						Scheme: secScheme.Scheme,
					}
				case "apiKey":
					scheme.APIKey = &openapi31.SecuritySchemeAPIKey{
						Name: secScheme.Name,
						In:   openapi31.SecuritySchemeAPIKeyIn(secScheme.In),
					}
				case "openIdConnect":
					scheme.Oidc = &openapi31.SecuritySchemeOidc{
						OpenIDConnectURL: secScheme.OpenIdConnectUrl,
					}
				case "oauth2":
					flows := openapi31.OauthFlows{}
					if secScheme.Flows.Implicit != nil {
						scopes := map[string]string{}
						for _, scope := range secScheme.Flows.Implicit.Scopes.AdditionalProperties {
							scopes[scope.Name] = scope.Value
						}
						flows.Implicit = &openapi31.OauthFlowsDefsImplicit{
							AuthorizationURL: secScheme.Flows.Implicit.AuthorizationUrl,
							RefreshURL:       &secScheme.Flows.Implicit.RefreshUrl,
							Scopes:           scopes,
						}
					}
					if secScheme.Flows.Password != nil {
						scopes := map[string]string{}
						for _, scope := range secScheme.Flows.Password.Scopes.AdditionalProperties {
							scopes[scope.Name] = scope.Value
						}
						flows.Password = &openapi31.OauthFlowsDefsPassword{
							TokenURL:   secScheme.Flows.Password.TokenUrl,
							RefreshURL: &secScheme.Flows.Password.RefreshUrl,
							Scopes:     scopes,
						}
					}
					if secScheme.Flows.ClientCredentials != nil {
						scopes := map[string]string{}
						for _, scope := range secScheme.Flows.ClientCredentials.Scopes.AdditionalProperties {
							scopes[scope.Name] = scope.Value
						}
						flows.ClientCredentials = &openapi31.OauthFlowsDefsClientCredentials{
							TokenURL:   secScheme.Flows.ClientCredentials.TokenUrl,
							RefreshURL: &secScheme.Flows.ClientCredentials.RefreshUrl,
							Scopes:     scopes,
						}
					}
					if secScheme.Flows.AuthorizationCode != nil {
						scopes := map[string]string{}
						for _, scope := range secScheme.Flows.AuthorizationCode.Scopes.AdditionalProperties {
							scopes[scope.Name] = scope.Value
						}
						flows.AuthorizationCode = &openapi31.OauthFlowsDefsAuthorizationCode{
							AuthorizationURL: secScheme.Flows.AuthorizationCode.AuthorizationUrl,
							TokenURL:         secScheme.Flows.AuthorizationCode.TokenUrl,
							RefreshURL:       &secScheme.Flows.AuthorizationCode.RefreshUrl,
							Scopes:           scopes,
						}
					}
					scheme.Oauth2 = &openapi31.SecuritySchemeOauth2{Flows: flows}
				default:
					continue
				}
				item.SecurityScheme = scheme
			}
			secSchemas[addProp.Name] = item
		}
		spec.Components.SecuritySchemes = secSchemas
	}
}

func addExternalDocsToSpec(spec *openapi31.Spec, externalDocs *goa3.ExternalDocs) {
	if externalDocs == nil {
		return
	}

	spec.ExternalDocs = &openapi31.ExternalDocumentation{
		Description: &externalDocs.Description,
		URL:         externalDocs.Url,
	}
}

func addTagsToSpec(spec *openapi31.Spec, tags []*goa3.Tag) {
	if len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		var extDoc *openapi31.ExternalDocumentation
		if tag.ExternalDocs != nil {
			extDoc = &openapi31.ExternalDocumentation{
				Description: &tag.Description,
				URL:         tag.ExternalDocs.Url,
			}
		}
		spec.Tags = append(spec.Tags, openapi31.Tag{
			Name:         tag.Name,
			Description:  &tag.Description,
			ExternalDocs: extDoc,
		})
	}
}
