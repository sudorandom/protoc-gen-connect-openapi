package googleapi

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
)

// AddSchemas adds the google.rpc.Status and google.protobuf.Any schemas used
// by the default error response of google.api.http-annotated methods. When
// WithGoogleErrorDetail is enabled, google.rpc.Status details are typed with
// the standard error detail messages.
func AddSchemas(opts options.Options, doc *v3.Document) {
	components := doc.Components

	if _, ok := components.Schemas.Get("google.protobuf.Any"); !ok {
		anyProps := orderedmap.New[string, *base.SchemaProxy]()
		anyProps.Set("@type", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"string"},
			Description: "A URL/resource name that uniquely identifies the type of the serialized message.",
		}))
		components.Schemas.Set("google.protobuf.Any", base.CreateSchemaProxy(&base.Schema{
			Type:                 []string{"object"},
			Description:          "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message.",
			Properties:           anyProps,
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
		}))
	}

	if _, ok := components.Schemas.Get("google.rpc.Status"); !ok {
		statusProps := orderedmap.New[string, *base.SchemaProxy]()
		statusProps.Set("code", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"integer"},
			Format:      "int32",
			Description: "The status code, which should be an enum value of google.rpc.Code.",
		}))
		statusProps.Set("message", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"string"},
			Description: "A developer-facing error message.",
		}))
		statusProps.Set("details", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"array"},
			Description: "A list of messages that carry the error details.",
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: statusDetailSchema(opts, components),
			},
		}))
		components.Schemas.Set("google.rpc.Status", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"object"},
			Description: "The Status type defines a logical error model suitable for gRPC and REST APIs.",
			Properties:  statusProps,
		}))
	}
}

// statusDetailSchema returns the schema for an entry in google.rpc.Status
// details. With WithGoogleErrorDetail, each standard error detail message is a
// variant keyed by its "@type"; anything else still matches google.protobuf.Any.
func statusDetailSchema(opts options.Options, components *v3.Components) *base.SchemaProxy {
	anyRef := base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Any")
	if !opts.WithGoogleErrorDetail {
		return anyRef
	}
	var variants []*base.SchemaProxy
	detailSchemas := util.GoogleRPCErrorDetailSchemas(opts)
	for pair := detailSchemas.First(); pair != nil; pair = pair.Next() {
		name := pair.Key()
		if _, ok := components.Schemas.Get(name); !ok {
			components.Schemas.Set(name, pair.Value())
		}
		typeProps := orderedmap.New[string, *base.SchemaProxy]()
		typeProps.Set("@type", base.CreateSchemaProxy(&base.Schema{
			Type:  []string{"string"},
			Const: utils.CreateStringNode("type.googleapis.com/" + name),
		}))
		variants = append(variants, base.CreateSchemaProxy(&base.Schema{
			Title:      pair.Value().Schema().Title,
			AllOf:      []*base.SchemaProxy{base.CreateSchemaProxyRef("#/components/schemas/" + name)},
			Properties: typeProps,
			Required:   []string{"@type"},
		}))
	}
	// anyOf rather than oneOf: every typed variant also matches
	// google.protobuf.Any.
	return base.CreateSchemaProxy(&base.Schema{AnyOf: append(variants, anyRef)})
}
