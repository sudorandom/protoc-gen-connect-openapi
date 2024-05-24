package gnostic

import (
	"strconv"

	goa3 "github.com/google/gnostic/openapiv3"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

func toServers(servers []*goa3.Server) []*v3.Server {
	result := make([]*v3.Server, len(servers))
	for i, server := range servers {
		result[i] = &v3.Server{
			URL:         server.Url,
			Description: server.Description,
			Variables:   toVariables(server.Variables),
		}
	}
	return result
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

//gocyclo:ignore
func toSecuritySchemes(components *goa3.Components) *orderedmap.Map[string, *v3.SecurityScheme] {
	if components == nil || components.SecuritySchemes == nil {
		return nil
	}

	secSchemas := orderedmap.New[string, *v3.SecurityScheme]()
	for _, addProp := range components.SecuritySchemes.AdditionalProperties {
		secScheme := addProp.Value.GetSecurityScheme()
		if secScheme != nil {
			scheme := &v3.SecurityScheme{}
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
					for _, scope := range secScheme.Flows.Password.Scopes.AdditionalProperties {
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
					for _, scope := range secScheme.Flows.Password.Scopes.AdditionalProperties {
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
	schema := &base.Schema{
		Title:         s.Title,
		Description:   s.Description,
		Default:       toDefault(s.Default),
		ReadOnly:      &s.ReadOnly,
		MultipleOf:    &s.MultipleOf,
		MaxLength:     &s.MaxLength,
		MinLength:     &s.MinLength,
		Pattern:       s.Pattern,
		MaxItems:      &s.MaxItems,
		MinItems:      &s.MinItems,
		UniqueItems:   &s.UniqueItems,
		MaxProperties: &s.MaxProperties,
		MinProperties: &s.MinProperties,
		Required:      s.Required,
		Format:        s.Format,
	}
	if s.Type != "" {
		schema.Type = []string{s.Type}
	}
	if s.ExclusiveMaximum {
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{B: s.Maximum}
	} else {
		schema.Maximum = &s.Maximum
	}
	if s.ExclusiveMinimum {
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{B: s.Minimum}
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

func toDefault(dt *goa3.DefaultType) *yaml.Node {
	if dt == nil {
		return nil
	}
	switch dt.GetOneof().(type) {
	case *goa3.DefaultType_Number:
		return &yaml.Node{Value: strconv.FormatFloat(dt.GetNumber(), 'f', -1, 64)}
	case *goa3.DefaultType_String_:
		return &yaml.Node{Value: dt.GetString_()}
	case *goa3.DefaultType_Boolean:
		if dt.GetBoolean() {
			return &yaml.Node{Value: "true"}
		}
		return &yaml.Node{Value: "false"}
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
