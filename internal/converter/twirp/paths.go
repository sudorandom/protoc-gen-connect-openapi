package twirp

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

const twirpPathPrefix = "/twirp"

func MakePathItems(opts options.Options, service protoreflect.ServiceDescriptor, method protoreflect.MethodDescriptor) *orderedmap.Map[string, *v3.PathItem] {
	paths := orderedmap.New[string, *v3.PathItem]()
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
	content := orderedmap.New[string, *v3.MediaType]()
	for contentType := range opts.ContentTypes {
		content.Set(contentType, &v3.MediaType{
			Schema: base.CreateSchemaProxyRef("#/components/schemas/" + string(message.FullName())),
		})
	}
	return &v3.RequestBody{
		Content: content,
	}
}

func makeResponses(opts options.Options, message protoreflect.MessageDescriptor) *v3.Responses {
	content := orderedmap.New[string, *v3.MediaType]()
	for contentType := range opts.ContentTypes {
		content.Set(contentType, &v3.MediaType{
			Schema: base.CreateSchemaProxyRef("#/components/schemas/" + string(message.FullName())),
		})
	}
	codes := orderedmap.New[string, *v3.Response]()
	codes.Set("200", &v3.Response{
		Description: "OK",
		Content:     content,
	})
	// Twirp errors are special. They are always JSON.
	errorContent := orderedmap.New[string, *v3.MediaType]()
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
	errorContent.Set("application/json", &v3.MediaType{
		Schema: base.CreateSchemaProxy(&base.Schema{
			Type:       []string{"object"},
			Properties: errorSchemaProperties,
		}),
	})
	codes.Set("default", &v3.Response{
		Description: "Error",
		Content:     errorContent,
	})
	return &v3.Responses{
		Codes: codes,
	}
}
