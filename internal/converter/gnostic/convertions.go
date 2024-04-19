package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/jsonschema-go"
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

//gocyclo:ignore
func toSecuritySchemes(components *goa3.Components) map[string]openapi31.SecuritySchemeOrReference {
	if components == nil || components.SecuritySchemes == nil {
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

func toSchemaOrBools(items []*goa3.SchemaOrReference) []jsonschema.SchemaOrBool {
	result := make([]jsonschema.SchemaOrBool, len(items))
	for i, s := range items {
		result[i] = toSchemaOrBool(s)
	}
	return result
}

func toSchemaOrBool(s *goa3.SchemaOrReference) jsonschema.SchemaOrBool {
	sOrB := jsonschema.SchemaOrBool{}
	if ref := s.GetReference(); ref != nil {
		sOrB.TypeObject = &jsonschema.Schema{
			Ref:         &ref.XRef,
			Description: &ref.Description,
		}
	} else if schema := s.GetSchema(); schema != nil {
		sOrB.TypeObject = toSchema(schema)
	}
	return sOrB
}

func toSchemaOrBoolMap(items []*goa3.NamedSchemaOrReference) map[string]jsonschema.SchemaOrBool {
	m := make(map[string]jsonschema.SchemaOrBool, len(items))
	for _, item := range items {
		m[item.Name] = toSchemaOrBool(item.Value)
	}
	return m
}

func toSchema(s *goa3.Schema) *jsonschema.Schema {
	schema := &jsonschema.Schema{
		Title:         &s.Title,
		Description:   &s.Description,
		Default:       toDefault(s.Default),
		ReadOnly:      &s.ReadOnly,
		MultipleOf:    &s.MultipleOf,
		MaxLength:     &s.MaxLength,
		MinLength:     s.MinLength,
		Pattern:       &s.Pattern,
		MaxItems:      &s.MaxItems,
		MinItems:      s.MinItems,
		UniqueItems:   &s.UniqueItems,
		MaxProperties: &s.MaxProperties,
		MinProperties: s.MinProperties,
		Required:      s.Required,
		Format:        &s.Format,
	}
	if s.Type != "" {
		t := jsonschema.SimpleType(s.Type)
		schema.Type = &jsonschema.Type{SimpleTypes: &t}
	}

	if s.ExclusiveMaximum {
		schema.ExclusiveMaximum = &s.Maximum
	} else {
		schema.Maximum = &s.Maximum
	}
	if s.ExclusiveMinimum {
		schema.ExclusiveMinimum = &s.Minimum
	} else {
		schema.Minimum = &s.Minimum
	}

	// Not Supported:
	// Items
	// Contains
	// AdditionalProperties
	// Definitions
	// Properties
	// PatternProperties
	// Dependencies
	// PropertyNames
	// Enum
	// If
	// Then
	// Else
	// AllOf
	// AnyOf
	// OneOf
	// Not
	// Parent
	return schema
}

func toDefault(dt *goa3.DefaultType) *interface{} {
	if dt == nil {
		return nil
	}
	var v interface{}
	switch dt.GetOneof().(type) {
	case *goa3.DefaultType_Number:
		v = dt.GetNumber()
	case *goa3.DefaultType_String_:
		v = dt.GetString_()
	case *goa3.DefaultType_Boolean:
		v = dt.GetBoolean()
	}
	return &v
}

func toAdditionalPropertiesItem(item *goa3.AdditionalPropertiesItem) *jsonschema.SchemaOrBool {
	switch v := item.Oneof.(type) {
	case *goa3.AdditionalPropertiesItem_SchemaOrReference:
		vv := toSchemaOrBool(v.SchemaOrReference)
		return &vv
	case *goa3.AdditionalPropertiesItem_Boolean:
		return &jsonschema.SchemaOrBool{TypeBoolean: &v.Boolean}
	}
	return nil
}
