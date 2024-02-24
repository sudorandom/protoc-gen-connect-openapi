package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/openapi-go/openapi31"
)

func toServers(servers []*goa3.Server) []openapi31.Server {
	result := make([]openapi31.Server, len(servers))
	for i, server := range servers {
		result[i] = openapi31.Server{
			URL:         server.Url,
			Description: &server.Description,
			Variables:   toVariables(server.Variables),
		}
	}
	return result
}

func toVariables(variables *goa3.ServerVariables) map[string]openapi31.ServerVariable {
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

func toSecurityRequirements(securityReq []*goa3.SecurityRequirement) []map[string][]string {
	result := make([]map[string][]string, len(securityReq))
	for i, req := range securityReq {
		item := map[string][]string{}
		for _, prop := range req.AdditionalProperties {
			item[prop.Name] = prop.Value.Value
		}
		result[i] = item
	}
	return result
}

func toSecuritySchemes(components *goa3.Components) map[string]openapi31.SecuritySchemeOrReference {
	if components.SecuritySchemes == nil {
		return nil
	}

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

	return secSchemas
}

func toExternalDocs(externalDocs *goa3.ExternalDocs) *openapi31.ExternalDocumentation {
	if externalDocs == nil {
		return nil
	}

	return &openapi31.ExternalDocumentation{
		Description: &externalDocs.Description,
		URL:         externalDocs.Url,
	}
}

func toTags(tags []*goa3.Tag) []openapi31.Tag {
	if len(tags) == 0 {
		return nil
	}

	result := make([]openapi31.Tag, len(tags))
	for i, tag := range tags {
		var extDoc *openapi31.ExternalDocumentation
		if tag.ExternalDocs != nil {
			extDoc = &openapi31.ExternalDocumentation{
				Description: &tag.Description,
				URL:         tag.ExternalDocs.Url,
			}
		}
		result[i] = openapi31.Tag{
			Name:         tag.Name,
			Description:  &tag.Description,
			ExternalDocs: extDoc,
		}
	}
	return result
}
