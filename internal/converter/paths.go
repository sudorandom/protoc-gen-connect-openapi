package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func fileToPathItems(opts options.Options, fd protoreflect.FileDescriptor) (*orderedmap.Map[string, *v3.PathItem], error) {
	items := orderedmap.New[string, *v3.PathItem]()
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			pathItems := googleapi.MakePathItems(opts, method)
			for pair := pathItems.First(); pair != nil; pair = pair.Next() {
				path, item := pair.Key(), pair.Value()
				if existing, ok := items.Get(pair.Key()); !ok {
					items.Set(path, item)
				} else {
					if item.Get != nil {
						existing.Get = item.Get
					}
					if item.Post != nil {
						existing.Post = item.Post
					}
					if item.Delete != nil {
						existing.Delete = item.Delete
					}
					if item.Put != nil {
						existing.Put = item.Put
					}
					if item.Patch != nil {
						existing.Patch = item.Patch
					}
					items.Set(path, existing)
				}
			}
			// No google.api annotations for this method, so default to the ConnectRPC/gRPC path
			if pathItems == nil || pathItems.Len() == 0 {
				items.Set("/"+string(service.FullName())+"/"+string(method.Name()), methodToPathItem(opts, method))
			}
		}
	}

	return items, nil
}

func methodToOperaton(opts options.Options, method protoreflect.MethodDescriptor, returnGet bool) *v3.Operation {
	fd := method.ParentFile()
	service := method.Parent().(protoreflect.ServiceDescriptor)
	loc := fd.SourceLocations().ByDescriptor(method)
	op := &v3.Operation{
		Summary:     string(method.Name()),
		OperationId: string(method.FullName()),
		Deprecated:  util.IsMethodDeprecated(method),
		Tags:        []string{string(service.FullName())},
		Description: util.FormatComments(loc),
	}

	isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
	if isStreaming && !opts.WithStreaming {
		return nil
	}

	// Responses
	codeMap := orderedmap.New[string, *v3.Response]()
	outputId := util.FormatTypeRef(string(method.Output().FullName()))
	codeMap.Set("200", &v3.Response{
		Description: "Success",
		Content: util.MakeMediaTypes(
			opts,
			base.CreateSchemaProxyRef("#/components/schemas/"+outputId),
			false,
			isStreaming,
		),
	})
	op.Responses = &v3.Responses{
		Codes: codeMap,
		Default: &v3.Response{
			Description: "Error",
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/connect.error"),
				false,
				isStreaming,
			),
		},
	}

	op.Parameters = append(op.Parameters,
		&v3.Parameter{
			Name:     "Connect-Protocol-Version",
			In:       "header",
			Required: util.BoolPtr(true),
			Schema:   base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
		},
		&v3.Parameter{
			Name:   "Connect-Timeout-Ms",
			In:     "header",
			Schema: base.CreateSchemaProxyRef("#/components/schemas/connect-timeout-header"),
		},
	)

	// Request parameters
	inputId := util.FormatTypeRef(string(method.Input().FullName()))
	if returnGet {
		op.OperationId = op.OperationId + ".get"
		op.Parameters = append(op.Parameters,
			&v3.Parameter{
				Name: "message",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/"+util.FormatTypeRef(inputId)),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "encoding",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/encoding"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "base64",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/base64"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "compression",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/compression"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "connect",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
					true,
					isStreaming),
			},
		)
	} else {
		op.RequestBody = &v3.RequestBody{
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/"+inputId),
				true,
				isStreaming,
			),
			Required: util.BoolPtr(true),
		}
	}

	return op
}

func methodToPathItem(opts options.Options, method protoreflect.MethodDescriptor) *v3.PathItem {
	hasGetSupport := methodHasGet(opts, method)
	item := &v3.PathItem{}
	if hasGetSupport {
		item.Get = methodToOperaton(opts, method, true)
	}
	item.Post = methodToOperaton(opts, method, false)
	item = gnostic.PathItemWithMethodAnnotations(item, method)

	return item
}

func methodHasGet(opts options.Options, method protoreflect.MethodDescriptor) bool {
	if !opts.AllowGET {
		return false
	}

	if method.IsStreamingClient() || method.IsStreamingServer() {
		return false
	}

	options := method.Options().(*descriptorpb.MethodOptions)
	return options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS
}
