package connectrpc

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
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

func newGoogleRPCErrorDetailSchemas() *orderedmap.Map[string, *base.SchemaProxy] {
	schemas := orderedmap.New[string, *base.SchemaProxy]()

	// ErrorInfo
	errorInfoProps := orderedmap.New[string, *base.SchemaProxy]()
	errorInfoProps.Set("reason", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The reason of the error. This is a constant value that identifies the proximate cause of the error. Error reasons are unique within a particular domain of errors. This should be at most 63 characters and match a regular expression of `[A-Z][A-Z0-9_]+[A-Z0-9]`, which represents UPPER_SNAKE_CASE.",
	}))
	errorInfoProps.Set("domain", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: `The logical grouping to which the "reason" belongs. The error domain is typically the registered service name of the tool or product that generates the error. Example: "pubsub.googleapis.com". If the error is generated by some common infrastructure, the error domain must be a globally unique value that identifies the infrastructure. For Google API infrastructure, the error domain is "googleapis.com".`,
	}))
	errorInfoProps.Set("metadata", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"string"},
			}),
		},
		Description: `Additional structured details about this error. Keys must match a regular expression of ` + "`[a-z][a-zA-Z0-9-_]+`" + ` but should ideally be lowerCamelCase. Also, they must be limited to 64 characters in length. When identifying the current value of an exceeded limit, the units should be contained in the key, not the value.  For example, rather than ` + "`{\"instanceLimit\": \"100/request\"}`" + `, should be returned as, ` + "`{\"instanceLimitPerRequest\": \"100\"}`" + `, if the client exceeds the number of instances that can be created in a single (batch) request.`,
	}))
	schemas.Set("google.rpc.ErrorInfo", base.CreateSchemaProxy(&base.Schema{
		Title:       "ErrorInfo",
		Type:        []string{"object"},
		Properties:  errorInfoProps,
		Description: "Describes the cause of the error with structured details.",
	}))

	// RetryInfo
	retryInfoProps := orderedmap.New[string, *base.SchemaProxy]()
	retryInfoProps.Set("retry_delay", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Format:      "duration",
		Description: "Clients consuming this error should wait at least this long before retrying. The value is a duration string, following the protobuf Duration format (e.g., \"1.5s\" for one and a half seconds).",
	}))
	schemas.Set("google.rpc.RetryInfo", base.CreateSchemaProxy(&base.Schema{
		Title:       "RetryInfo",
		Type:        []string{"object"},
		Properties:  retryInfoProps,
		Description: "Describes when the clients can retry a failed request.",
	}))

	// DebugInfo
	debugInfoProps := orderedmap.New[string, *base.SchemaProxy]()
	debugInfoProps.Set("stack_entries", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type: []string{"string"},
			}),
		},
	}))
	debugInfoProps.Set("detail", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"string"},
	}))
	schemas.Set("google.rpc.DebugInfo", base.CreateSchemaProxy(&base.Schema{
		Title:       "DebugInfo",
		Type:        []string{"object"},
		Properties:  debugInfoProps,
		Description: "Contains debugging information.",
	}))

	// QuotaFailure
	quotaFailureViolationProps := orderedmap.New[string, *base.SchemaProxy]()
	quotaFailureViolationProps.Set("subject", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The subject on which the quota check failed.",
	}))
	quotaFailureViolationProps.Set("description", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A description of how the quota check failed.",
	}))
	quotaFailureProps := orderedmap.New[string, *base.SchemaProxy]()
	quotaFailureProps.Set("violations", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: quotaFailureViolationProps,
			}),
		},
	}))
	schemas.Set("google.rpc.QuotaFailure", base.CreateSchemaProxy(&base.Schema{
		Title:       "QuotaFailure",
		Type:        []string{"object"},
		Properties:  quotaFailureProps,
		Description: "Describes how a quota check failed.",
	}))

	// PreconditionFailure
	preconditionFailureViolationProps := orderedmap.New[string, *base.SchemaProxy]()
	preconditionFailureViolationProps.Set("type", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The type of PreconditionFailure. We recommend using a service-specific enum type to define the supported precondition violation types.",
	}))
	preconditionFailureViolationProps.Set("subject", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The subject, relative to the type, that failed the precondition.",
	}))
	preconditionFailureViolationProps.Set("description", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A description of how the precondition failed.",
	}))
	preconditionFailureProps := orderedmap.New[string, *base.SchemaProxy]()
	preconditionFailureProps.Set("violations", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: preconditionFailureViolationProps,
			}),
		},
	}))
	schemas.Set("google.rpc.PreconditionFailure", base.CreateSchemaProxy(&base.Schema{
		Title:       "PreconditionFailure",
		Type:        []string{"object"},
		Properties:  preconditionFailureProps,
		Description: "A message type used to describe a failed precondition.",
	}))

	// BadRequest
	badRequestFieldViolationProps := orderedmap.New[string, *base.SchemaProxy]()
	badRequestFieldViolationProps.Set("field", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A path that leads to a field in the request body.",
	}))
	badRequestFieldViolationProps.Set("description", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A description of why the request element is bad.",
	}))
	badRequestProps := orderedmap.New[string, *base.SchemaProxy]()
	badRequestProps.Set("field_violations", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: badRequestFieldViolationProps,
			}),
		},
	}))
	schemas.Set("google.rpc.BadRequest", base.CreateSchemaProxy(&base.Schema{
		Title:       "BadRequest",
		Type:        []string{"object"},
		Properties:  badRequestProps,
		Description: "A message type used to describe a bad request.",
	}))

	// RequestInfo
	requestInfoProps := orderedmap.New[string, *base.SchemaProxy]()
	requestInfoProps.Set("request_id", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "An opaque string that could be used by the client to trace the request.",
	}))
	requestInfoProps.Set("serving_data", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The serving data that were used to process the request.",
	}))
	schemas.Set("google.rpc.RequestInfo", base.CreateSchemaProxy(&base.Schema{
		Title:       "RequestInfo",
		Type:        []string{"object"},
		Properties:  requestInfoProps,
		Description: "Contains metadata about the request that clients can attach when filing a bug or providing other forms of feedback.",
	}))

	// ResourceInfo
	resourceInfoProps := orderedmap.New[string, *base.SchemaProxy]()
	resourceInfoProps.Set("resource_type", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: `A name for the type of resource being accessed, e.g. "sql table", "cloud storage bucket", etc.`,
	}))
	resourceInfoProps.Set("resource_name", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The name of the resource being accessed.",
	}))
	resourceInfoProps.Set("owner", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The owner of the resource (optional).",
	}))
	resourceInfoProps.Set("description", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A description of the resource (optional).",
	}))
	schemas.Set("google.rpc.ResourceInfo", base.CreateSchemaProxy(&base.Schema{
		Title:       "ResourceInfo",
		Type:        []string{"object"},
		Properties:  resourceInfoProps,
		Description: "Describes the resource that is being accessed.",
	}))

	// Help
	helpLinkProps := orderedmap.New[string, *base.SchemaProxy]()
	helpLinkProps.Set("description", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "A description of the link.",
	}))
	helpLinkProps.Set("url", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The URL of the link.",
	}))
	helpProps := orderedmap.New[string, *base.SchemaProxy]()
	helpProps.Set("links", base.CreateSchemaProxy(&base.Schema{
		Type: []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: helpLinkProps,
			}),
		},
	}))
	schemas.Set("google.rpc.Help", base.CreateSchemaProxy(&base.Schema{
		Title:       "Help",
		Type:        []string{"object"},
		Properties:  helpProps,
		Description: "Provides links to documentation or for performing an out-of-band action.",
	}))

	// LocalizedMessage
	localizedMessageProps := orderedmap.New[string, *base.SchemaProxy]()
	localizedMessageProps.Set("locale", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The locale used following the specification defined at http://www.rfc-editor.org/rfc/bcp/bcp47.txt.",
	}))
	localizedMessageProps.Set("message", base.CreateSchemaProxy(&base.Schema{
		Type:        []string{"string"},
		Description: "The localized message in the locale.",
	}))
	schemas.Set("google.rpc.LocalizedMessage", base.CreateSchemaProxy(&base.Schema{
		Title:       "LocalizedMessage",
		Type:        []string{"object"},
		Properties:  localizedMessageProps,
		Description: "A message type used to provide a localized message.",
	}))

	return schemas
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

	if _, ok := components.Schemas.Get("base64"); !ok {
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
