package converter

import (
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func fileToPathItems(opts Options, fd protoreflect.FileDescriptor) (map[string]openapi31.PathItem, error) {
	items := map[string]openapi31.PathItem{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			pathItems := googleapi.MakePathItems(method)
			for path, item := range pathItems {
				if existing, ok := items[path]; !ok {
					items[path] = item
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
					if existing.MapOfAnything == nil {
						existing.MapOfAnything = map[string]interface{}{}
					}
					for k, v := range item.MapOfAnything {
						existing.MapOfAnything[k] = v
					}
					items[path] = existing
				}
			}
			if len(pathItems) == 0 {
				// Default ConnectRPC/gRPC path
				items["/"+string(service.FullName())+"/"+string(method.Name())] = methodToPathItem(opts, method)
			}
		}
	}

	return items, nil
}

func methodToPathItem(opts Options, method protoreflect.MethodDescriptor) openapi31.PathItem {
	fd := method.ParentFile()
	service := method.Parent().(protoreflect.ServiceDescriptor)
	op := &openapi31.Operation{
		Deprecated: util.IsMethodDeprecated(method),
	}
	op.WithTags(string(service.FullName()))
	loc := fd.SourceLocations().ByDescriptor(method)
	op.WithDescription(util.FormatComments(loc))

	hasGetSupport := methodHasGet(opts, method)

	// Responses
	responses := openapi31.Responses{
		Default: &openapi31.ResponseOrReference{
			Reference: &openapi31.Reference{Ref: "#/components/responses/connect.error"},
		},
	}
	if !util.IsEmpty(method.Output()) {
		id := util.FormatTypeRef(string(method.FullName() + "." + method.Output().FullName()))
		responses.WithMapOfResponseOrReferenceValuesItem("200", openapi31.ResponseOrReference{
			Reference: &openapi31.Reference{Ref: "#/components/responses/" + id},
		})
	}
	op.WithResponses(responses)

	// Request parameters
	item := openapi31.PathItem{}
	if !util.IsEmpty(method.Input()) {
		id := util.FormatTypeRef(string(method.FullName() + "." + method.Input().FullName()))
		if hasGetSupport {
			ref := "#/components/parameters/" + id
			op.Parameters = append(op.Parameters, openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: ref}})
		} else {
			op.WithRequestBody(openapi31.RequestBodyOrReference{
				Reference: &openapi31.Reference{Ref: "#/components/requestBodies/" + id},
			})
		}
	}

	if hasGetSupport {
		item.Get = op
		op.Parameters = append(op.Parameters,
			openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/encoding"}},
			openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/base64"}},
			openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/compression"}},
			openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/connect"}},
		)
	} else {
		item.Post = op
	}

	item = gnostic.PathItemWithMethodAnnotations(item, method)

	return item
}

func methodHasGet(opts Options, method protoreflect.MethodDescriptor) bool {
	if !opts.AllowGET {
		return false
	}

	if method.IsStreamingClient() || method.IsStreamingServer() {
		return false
	}

	options := method.Options().(*descriptorpb.MethodOptions)
	return options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS
}
