package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
)

func addConnectSchemas(opts options.Options, components *v3.Components) {
	if _, ok := components.Schemas.Get("connect-protocol-version"); !ok {
		components.Schemas.Set("connect-protocol-version", base.CreateSchemaProxy(&base.Schema{
			Title:       "Connect-Protocol-Version",
			Description: "Define the version of the Connect protocol",
			Type:        []string{"number"},
			Enum:        []*yaml.Node{utils.CreateIntNode("1")},
			Const:       utils.CreateIntNode("1"),
		}))
	}

	if _, ok := components.Schemas.Get("connect-timeout-header"); !ok {
		components.Schemas.Set("connect-timeout-header", base.CreateSchemaProxy(&base.Schema{
			Title:       "Connect-Timeout-Ms",
			Description: "Define the timeout, in ms",
			Type:        []string{"number"},
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
		connectAnyProps := orderedmap.New[string, *base.SchemaProxy]()
		connectAnyProps.Set("type", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"string"},
			Description: "A URL that acts as a globally unique identifier for the type of the serialized message. For example: `type.googleapis.com/google.rpc.ErrorInfo`. This is used to determine the schema of the data in the `value` field and is the discriminator for the `debug` field.",
		}))
		connectAnyProps.Set("value", base.CreateSchemaProxy(&base.Schema{
			Type:        []string{"string"},
			Format:      "binary",
			Description: "The Protobuf message, serialized as bytes and base64-encoded. The specific message type is identified by the `type` field.",
		}))

		errorDetailOptions := []*base.SchemaProxy{
			base.CreateSchemaProxy(&base.Schema{
				Title:                "Any",
				Description:          "Detailed error information.",
				Type:                 []string{"object"},
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
			}),
		}
		mapping := orderedmap.New[string, string]()
		if opts.WithGoogleErrorDetail {
			googleRPCSchemas := newGoogleRPCErrorDetailSchemas()
			for pair := googleRPCSchemas.First(); pair != nil; pair = pair.Next() {
				components.Schemas.Set(pair.Key(), pair.Value())
				errorDetailOptions = append(errorDetailOptions, base.CreateSchemaProxyRef("#/components/schemas/"+pair.Key()))
			}
			for pair := googleRPCSchemas.First(); pair != nil; pair = pair.Next() {
				// The key is the full type URL, the value is the schema reference
				mapping.Set("type.googleapis.com/"+pair.Key(), "#/components/schemas/"+pair.Key())
			}
		}

		// Now create the schema for the "debug" field with the discriminator
		debugSchema := base.CreateSchemaProxy(&base.Schema{
			Title:       "Debug",
			Description: `Deserialized error detail payload. The 'type' field indicates the schema. This field is for easier debugging and should not be relied upon for application logic.`,
			OneOf:       errorDetailOptions,
			Discriminator: &base.Discriminator{
				PropertyName: "type",
				Mapping:      mapping,
			},
		})
		connectAnyProps.Set("debug", debugSchema)

		components.Schemas.Set("connect.error_details.Any", base.CreateSchemaProxy(&base.Schema{
			Description:          "Contains an arbitrary serialized message along with a @type that describes the type of the serialized message, with an additional debug field for ConnectRPC error details.",
			Type:                 []string{"object"},
			Properties:           connectAnyProps,
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
		}))
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
		}))
	}

	if _, ok := components.Schemas.Get("bas64"); !ok {
		components.Schemas.Set("base64", base.CreateSchemaProxy(&base.Schema{
			Title:       "base64",
			Description: "Specifies if the message query param is base64 encoded, which may be required for binary data",
			Type:        []string{"boolean"},
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
		}))
	}

	if _, ok := components.Schemas.Get("connect"); !ok {
		components.Schemas.Set("connect", base.CreateSchemaProxy(&base.Schema{
			Title:       "connect",
			Description: "Define the version of the Connect protocol",
			Enum: []*yaml.Node{
				utils.CreateStringNode("v1"),
			},
		}))
	}
}
