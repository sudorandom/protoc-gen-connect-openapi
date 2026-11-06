package connectrpc

import (
	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func AddSchemas(opts options.Options, doc *v3.Document, method protoreflect.MethodDescriptor) {
	if methodHasGet(opts, method) {
		addConnectGetSchemas(doc.Components)
	}
	components := doc.Components
	if _, ok := components.Schemas.Get("connect-protocol-version"); !ok {
		components.Schemas.Set("connect-protocol-version", base.CreateSchemaProxy(&base.Schema{
			Title:       "Connect-Protocol-Version",
			Description: "Define the version of the Connect protocol",
			Type:        []string{"number"},
			Const:       utils.CreateIntNode("1"),
			Examples:    []*yaml.Node{utils.CreateIntNode("1")},
		}))
	}

	if _, ok := components.Schemas.Get("connect-timeout-header"); !ok {
		components.Schemas.Set("connect-timeout-header", base.CreateSchemaProxy(&base.Schema{
			Title:       "Connect-Timeout-Ms",
			Description: "Define the timeout, in ms",
			Type:        []string{"number"},
			Examples:    []*yaml.Node{utils.CreateIntNode("1000")},
		}))
	}

	if _, ok := components.Schemas.Get("ct.error"); !ok {
		connectErrorProps := orderedmap.New[string, *base.SchemaProxy]()
		connectErrorProps.Set("code", base.CreateSchemaProxy(&base.Schema{
			Description: "The status code, which should be an enum value of [google.rpc.Code][google.rpc.Code].",
			Type:        []string{"string"},
			Examples:    []*yaml.Node{utils.CreateStringNode("not_found")},
			Enum: []*yaml.Node{
				utils.CreateStringNode("canceled"),
				utils.CreateStringNode("unknown"),
				utils.CreateStringNode("invalid_argument"),
				utils.CreateStringNode("deadline_exceeded"),
				utils.CreateStringNode("not_found"),
				utils.CreateStringNode("already_exists"),
				utils.CreateStringNode("permission_denied"),
				utils.CreateStringNode("resource_exhausted"),
				utils.CreateStringNode("failed_precondition"),
				utils.CreateStringNode("aborted"),
				utils.CreateStringNode("out_of_range"),
				utils.CreateStringNode("unimplemented"),
				utils.CreateStringNode("internal"),
				utils.CreateStringNode("unavailable"),
				utils.CreateStringNode("data_loss"),
				utils.CreateStringNode("unauthenticated"),
			},
		}))
		connectErrorProps.Set("message", base.CreateSchemaProxy(&base.Schema{
			Description: "A developer-facing error message, which should be in English. Any user-facing error message should be localized and sent in the [google.rpc.Status.details][google.rpc.Status.details] field, or localized by the client.",
			Type:        []string{"string"},
		}))

		connectErrorProps.Set("details", base.CreateSchemaProxy(&base.Schema{
			Type: []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: base.CreateSchemaProxyRef("#/components/schemas/connect.error_details.Any"),
			},
			Description: "A list of messages that carry the error details. There is no limit on the number of messages.",
		}))
		components.Schemas.Set("connect.error", base.CreateSchemaProxy(&base.Schema{
			Title:                "Connect Error",
			Description:          `Error type returned by Connect: https://connectrpc.com/docs/go/errors/#http-representation`,
			Properties:           connectErrorProps,
			Type:                 []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
		}))
	}

	if _, ok := components.Schemas.Get("connect.error_details.Any"); !ok {
		if opts.WithGoogleErrorDetail {
			googleRPCSchemas := util.GoogleRPCErrorDetailSchemas()
			mapping := orderedmap.New[string, string]()
			oneOf := []*base.SchemaProxy{}

			for pair := googleRPCSchemas.First(); pair != nil; pair = pair.Next() {
				components.Schemas.Set(pair.Key(), pair.Value())

				detailSchemaName := "connect.error_details." + pair.Key()
				detailProps := orderedmap.New[string, *base.SchemaProxy]()
				detailProps.Set("type", base.CreateSchemaProxy(&base.Schema{
					Type:        []string{"string"},
					Description: "A URL that acts as a globally unique identifier for the type of the serialized message.",
				}))
				detailProps.Set("value", base.CreateSchemaProxy(&base.Schema{
					Type:        []string{"string"},
					Format:      "byte",
					Description: "The Protobuf message, serialized as bytes and base64-encoded. The specific message type is identified by the `type` field.",
				}))
				detailProps.Set("debug", base.CreateSchemaProxyRef("#/components/schemas/"+pair.Key()))

				components.Schemas.Set(detailSchemaName, base.CreateSchemaProxy(&base.Schema{
					Title:                pair.Key() + "Detail",
					Description:          "Connect error detail for " + pair.Key(),
					Type:                 []string{"object"},
					Properties:           detailProps,
					Required:             []string{"type"},
					AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
				}))

				oneOf = append(oneOf, base.CreateSchemaProxyRef("#/components/schemas/"+detailSchemaName))
				mapping.Set(pair.Key(), "#/components/schemas/"+detailSchemaName)
				mapping.Set("type.googleapis.com/"+pair.Key(), "#/components/schemas/"+detailSchemaName)
			}

			// Unknown / generic detail fallback
			unknownDetailProps := orderedmap.New[string, *base.SchemaProxy]()
			unknownDetailProps.Set("type", base.CreateSchemaProxy(&base.Schema{
				Type:        []string{"string"},
				Description: "A URL that acts as a globally unique identifier for the type of the serialized message.",
			}))
			unknownDetailProps.Set("value", base.CreateSchemaProxy(&base.Schema{
				Type:        []string{"string"},
				Format:      "byte",
				Description: "The Protobuf message, serialized as bytes and base64-encoded. The specific message type is identified by the `type` field.",
			}))
			unknownDetailProps.Set("debug", base.CreateSchemaProxy(&base.Schema{
				Title:                "Any",
				Description:          "Detailed error information.",
				Type:                 []string{"object"},
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
			}))
			components.Schemas.Set("connect.error_details.Unknown", base.CreateSchemaProxy(&base.Schema{
				Title:                "UnknownDetail",
				Description:          "Generic Connect error detail for unknown or custom message types.",
				Type:                 []string{"object"},
				Properties:           unknownDetailProps,
				Required:             []string{"type"},
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
			}))
			oneOf = append(oneOf, base.CreateSchemaProxyRef("#/components/schemas/connect.error_details.Unknown"))

			components.Schemas.Set("connect.error_details.Any", base.CreateSchemaProxy(&base.Schema{
				Title:       "ConnectErrorDetails",
				Description: "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message, with an additional debug field for ConnectRPC error details.",
				Type:        []string{"object"},
				Discriminator: &base.Discriminator{
					PropertyName: "type",
					Mapping:      mapping,
				},
				OneOf: oneOf,
			}))
		} else {
			connectAnyProps := orderedmap.New[string, *base.SchemaProxy]()
			connectAnyProps.Set("type", base.CreateSchemaProxy(&base.Schema{
				Type:        []string{"string"},
				Description: "A URL that acts as a globally unique identifier for the type of the serialized message. For example: `type.googleapis.com/google.rpc.ErrorInfo`.",
			}))
			connectAnyProps.Set("value", base.CreateSchemaProxy(&base.Schema{
				Type:        []string{"string"},
				Format:      "byte",
				Description: "The Protobuf message, serialized as bytes and base64-encoded. The specific message type is identified by the `type` field.",
			}))
			connectAnyProps.Set("debug", base.CreateSchemaProxy(&base.Schema{
				Title:                "Debug",
				Description:          "Deserialized error detail payload. The 'type' field indicates the schema. This field is for easier debugging and should not be relied upon for application logic.",
				Type:                 []string{"object"},
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
			}))

			components.Schemas.Set("connect.error_details.Any", base.CreateSchemaProxy(&base.Schema{
				Description:          "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message, with an additional debug field for ConnectRPC error details.",
				Type:                 []string{"object"},
				Properties:           connectAnyProps,
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
			}))
		}
	}
}

func addConnectGetSchemas(components *v3.Components) {
	if _, ok := components.Schemas.Get("encoding"); !ok {
		components.Schemas.Set("encoding", base.CreateSchemaProxy(&base.Schema{
			Title:       "encoding",
			Description: "Define which encoding or 'Message-Codec' to use",
			Enum: []*yaml.Node{
				utils.CreateStringNode("proto"),
				utils.CreateStringNode("json"),
			},
			Examples: []*yaml.Node{utils.CreateStringNode("json")},
		}))
	}

	if _, ok := components.Schemas.Get("base64"); !ok {
		components.Schemas.Set("base64", base.CreateSchemaProxy(&base.Schema{
			Title:       "base64",
			Description: "Specifies if the message query param is base64 encoded, which may be required for binary data",
			Type:        []string{"boolean"},
			Examples:    []*yaml.Node{utils.CreateBoolNode("false")},
		}))
	}

	if _, ok := components.Schemas.Get("compression"); !ok {
		components.Schemas.Set("compression", base.CreateSchemaProxy(&base.Schema{
			Title:       "compression",
			Description: "Which compression algorithm to use for this request",
			Enum: []*yaml.Node{
				utils.CreateStringNode("identity"),
				utils.CreateStringNode("gzip"),
				utils.CreateStringNode("br"),
			},
			Examples: []*yaml.Node{utils.CreateStringNode("gzip")},
		}))
	}

	if _, ok := components.Schemas.Get("connect"); !ok {
		components.Schemas.Set("connect", base.CreateSchemaProxy(&base.Schema{
			Title:       "connect",
			Description: "Define the version of the Connect protocol",
			Enum: []*yaml.Node{
				utils.CreateStringNode("v1"),
			},
			Examples: []*yaml.Node{utils.CreateStringNode("v1")},
		}))
	}
}
