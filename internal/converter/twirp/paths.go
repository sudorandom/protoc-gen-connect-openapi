package twirp

import (
	"fmt"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const twirpPathPrefix = "/twirp"

// twirpContentTypes are the encodings Twirp serves, in output order, keyed by
// the content-types option that enables them. Twirp's binary encoding uses
// application/protobuf rather than Connect's application/proto.
var twirpContentTypes = []struct{ option, mediaType string }{
	{option: "json", mediaType: "application/json"},
	{option: "proto", mediaType: "application/protobuf"},
}

func MakePathItems(opts options.Options, service protoreflect.ServiceDescriptor, method protoreflect.MethodDescriptor) *orderedmap.Map[string, *v3.PathItem] {
	paths := orderedmap.New[string, *v3.PathItem]()
	// Twirp only supports unary RPCs.
	if method.IsStreamingClient() || method.IsStreamingServer() {
		return paths
	}
	path := fmt.Sprintf("%s/%s/%s", twirpPathPrefix, service.FullName(), method.Name())
	paths.Set(path, MakePathItem(opts, method))
	return paths
}

func MakePathItem(opts options.Options, method protoreflect.MethodDescriptor) *v3.PathItem {
	return &v3.PathItem{
		Post: makeOperation(opts, method),
	}
}

func makeOperation(opts options.Options, method protoreflect.MethodDescriptor) *v3.Operation {
	return &v3.Operation{
		Tags:        []string{string(method.Parent().(protoreflect.ServiceDescriptor).FullName())},
		OperationId: string(method.FullName()),
		RequestBody: makeRequestBody(opts, method.Input()),
		Responses:   makeResponses(opts, method.Output()),
	}
}

func makeRequestBody(opts options.Options, message protoreflect.MessageDescriptor) *v3.RequestBody {
	content := makeContent(opts, message)
	return &v3.RequestBody{
		Content: content,
	}
}

func makeContent(opts options.Options, message protoreflect.MessageDescriptor) *orderedmap.Map[string, *v3.MediaType] {
	content := orderedmap.New[string, *v3.MediaType]()
	for _, contentType := range twirpContentTypes {
		if _, ok := opts.ContentTypes[contentType.option]; !ok {
			continue
		}
		content.Set(contentType.mediaType, &v3.MediaType{
			Schema: base.CreateSchemaProxyRef("#/components/schemas/" + string(message.FullName())),
		})
	}
	return content
}

func makeResponses(opts options.Options, message protoreflect.MessageDescriptor) *v3.Responses {
	content := makeContent(opts, message)
	codes := orderedmap.New[string, *v3.Response]()
	if !opts.DisableDefaultResponse {
		codes.Set("200", &v3.Response{
			Description: "OK",
			Content:     content,
		})
	}
	// Twirp errors are special. They are always JSON.
	errorContent := orderedmap.New[string, *v3.MediaType]()
	errorContent.Set("application/json", &v3.MediaType{
		Schema: base.CreateSchemaProxyRef("#/components/schemas/TwirpError"),
	})
	codes.Set("default", &v3.Response{
		Description: "Error",
		Content:     errorContent,
	})
	return &v3.Responses{
		Codes: codes,
	}
}

func AddSchemas(opts options.Options, doc *v3.Document, method protoreflect.MethodDescriptor) {
	components := doc.Components
	if _, ok := components.Schemas.Get("TwirpError"); !ok {
		errorSchemaProperties := orderedmap.New[string, *base.SchemaProxy]()
		errorSchemaProperties.Set("code", base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
			Enum: []*yaml.Node{
				utils.CreateStringNode("canceled"),
				utils.CreateStringNode("unknown"),
				utils.CreateStringNode("invalid_argument"),
				utils.CreateStringNode("malformed"),
				utils.CreateStringNode("deadline_exceeded"),
				utils.CreateStringNode("not_found"),
				utils.CreateStringNode("bad_route"),
				utils.CreateStringNode("already_exists"),
				utils.CreateStringNode("permission_denied"),
				utils.CreateStringNode("unauthenticated"),
				utils.CreateStringNode("resource_exhausted"),
				utils.CreateStringNode("failed_precondition"),
				utils.CreateStringNode("aborted"),
				utils.CreateStringNode("out_of_range"),
				utils.CreateStringNode("unimplemented"),
				utils.CreateStringNode("internal"),
				utils.CreateStringNode("unavailable"),
				utils.CreateStringNode("dataloss"),
			},
		}))
		errorSchemaProperties.Set("msg", base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}))
		errorSchemaProperties.Set("meta", base.CreateSchemaProxy(&base.Schema{
			Type:                 []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}})},
		}))
		components.Schemas.Set("TwirpError", base.CreateSchemaProxy(&base.Schema{
			Type:       []string{"object"},
			Properties: errorSchemaProperties,
		}))
	}
}
