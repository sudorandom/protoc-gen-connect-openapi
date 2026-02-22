package googleapi

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
)

// AddSchemas adds google.rpc.Status and google.protobuf.Any schemas to the
// OpenAPI document components when WithGoogleErrorDetail is enabled. This gives
// REST client generators typed error handling for google.api.http-annotated
// methods.
func AddSchemas(opts options.Options, doc *v3.Document) {
	if !opts.WithGoogleErrorDetail {
		return
	}
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
				A: base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Any"),
			},
		}))
		components.Schemas.Set("google.rpc.Status", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"object"},
			Description: "The Status type defines a logical error model suitable for gRPC and REST APIs.",
			Properties:  statusProps,
		}))
	}
}
