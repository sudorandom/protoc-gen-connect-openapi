package converter

import (
	"log/slog"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func fileToComponents(opts options.Options, fd protoreflect.FileDescriptor) (*highv3.Components, error) {
	// Add schema from messages/enums
	components := &highv3.Components{
		Schemas:         orderedmap.New[string, *base.SchemaProxy](),
		Responses:       orderedmap.New[string, *v3.Response](),
		Parameters:      orderedmap.New[string, *v3.Parameter](),
		Examples:        orderedmap.New[string, *base.Example](),
		RequestBodies:   orderedmap.New[string, *v3.RequestBody](),
		Headers:         orderedmap.New[string, *v3.Header](),
		SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
		Links:           orderedmap.New[string, *v3.Link](),
		Callbacks:       orderedmap.New[string, *v3.Callback](),
		Extensions:      orderedmap.New[string, *yaml.Node](),
	}
	st := NewState(opts)
	slog.Debug("start collection")
	st.CollectFile(fd)
	slog.Debug("collection complete", slog.String("file", string(fd.Name())), slog.Int("messages", len(st.Messages)), slog.Int("enum", len(st.Enums)))
	components.Schemas = stateToSchema(st)

	hasGetRequests := false

	// Add requestBodies and responses for methods
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			hasGet := methodHasGet(opts, method)
			if hasGet {
				hasGetRequests = true
			}
		}
	}

	if hasGetRequests {
		components.Parameters.Set("encoding", &v3.Parameter{
			Name:    "encoding",
			In:      "query",
			Content: util.MakeMediaTypes(opts, "#/components/schemas/encoding", true, false),
		})
		components.Schemas.Set("encoding", base.CreateSchemaProxy(&base.Schema{
			Title:       "encoding",
			Description: "Define which encoding or 'Message-Codec' to use",
			Enum: []*yaml.Node{
				{Value: "proto"},
				{Value: "json"},
			},
		}))

		components.Parameters.Set("base64", &v3.Parameter{
			Name:    "base64",
			In:      "query",
			Content: util.MakeMediaTypes(opts, "#/components/schemas/base64", true, false),
		})
		components.Schemas.Set("base64", base.CreateSchemaProxy(&base.Schema{
			Title:       "base64",
			Description: "Specifies if the message query param is base64 encoded, which may be required for binary data",
			Type:        []string{"boolean"},
		}))

		components.Parameters.Set("compression", &v3.Parameter{
			Name:    "compression",
			In:      "query",
			Content: util.MakeMediaTypes(opts, "#/components/schemas/compression", true, false),
		})
		components.Schemas.Set("compression", base.CreateSchemaProxy(&base.Schema{
			Title:       "compression",
			Description: "Which compression algorithm to use for this request",
			Enum: []*yaml.Node{
				{Value: "identity"},
				{Value: "gzip"},
				{Value: "br"},
				{Value: "gzip"},
			},
		}))

		components.Parameters.Set("connect", &v3.Parameter{
			Name:    "connect",
			In:      "query",
			Content: util.MakeMediaTypes(opts, "#/components/schemas/connect", true, false),
		})
		components.Schemas.Set("connect", base.CreateSchemaProxy(&base.Schema{
			Title:       "connect",
			Description: "Which version of connect to use.",
			Enum: []*yaml.Node{
				{Value: "1"},
			},
		}))
	}
	connectErrorProps := orderedmap.New[string, *base.SchemaProxy]()
	connectErrorProps.Set("code", base.CreateSchemaProxy(&base.Schema{
		Description: "The status code, which should be an enum value of [google.rpc.Code][google.rpc.Code].",
		Type:        []string{"string"},
		Examples:    []*yaml.Node{{Value: "CodeNotFound"}},
		Enum: []*yaml.Node{
			{Value: "CodeCanceled"},
			{Value: "CodeUnknown"},
			{Value: "CodeInvalidArgument"},
			{Value: "CodeDeadlineExceeded"},
			{Value: "CodeNotFound"},
			{Value: "CodeAlreadyExists"},
			{Value: "CodePermissionDenied"},
			{Value: "CodeResourceExhausted"},
			{Value: "CodeFailedPrecondition"},
			{Value: "CodeAborted"},
			{Value: "CodeOutOfRange"},
			{Value: "CodeInternal"},
			{Value: "CodeUnavailable"},
			{Value: "CodeDataLoss"},
			{Value: "CodeUnauthenticated"},
		},
	}))
	connectErrorProps.Set("message", base.CreateSchemaProxy(&base.Schema{
		Description: "A developer-facing error message, which should be in English. Any user-facing error message should be localized and sent in the [google.rpc.Status.details][google.rpc.Status.details] field, or localized by the client.",
		Type:        []string{"string"},
	}))
	connectErrorProps.Set("detail", base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Any"))
	components.Schemas.Set("connect.error", base.CreateSchemaProxy(&base.Schema{
		Title:                "Connect Error",
		Description:          `Error type returned by Connect: https://connectrpc.com/docs/go/errors/#http-representation`,
		Properties:           connectErrorProps,
		Type:                 []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
	}))
	anyPair := util.NewGoogleAny()
	components.Schemas.Set(anyPair.ID, base.CreateSchemaProxy(anyPair.Schema))

	return components, nil
}
