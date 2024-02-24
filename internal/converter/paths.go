package converter

import (
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func fileToPathItems(fd protoreflect.FileDescriptor) (map[string]openapi31.PathItem, error) {
	items := map[string]openapi31.PathItem{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			op := &openapi31.Operation{}
			op.WithTags(string(service.FullName()))
			loc := fd.SourceLocations().ByDescriptor(method)
			op.WithDescription(formatComments(loc))

			hasGetSupport := methodHasGet(method)

			// Request parameters
			parameters := []openapi31.ParameterOrReference{}
			item := openapi31.PathItem{}
			if !IsEmpty(method.Input()) {
				id := formatTypeRef(string(method.FullName() + "." + method.Input().FullName()))
				if hasGetSupport {
					ref := "#/components/parameters/" + id
					parameters = append(parameters, openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: ref}})
				} else {
					op.WithRequestBody(openapi31.RequestBodyOrReference{
						Reference: &openapi31.Reference{
							Ref: "#/components/requestBodies/" + id,
						},
					})
				}
			}

			// Responses
			responses := openapi31.Responses{
				Default: &openapi31.ResponseOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/responses/connect.error",
					},
				},
			}
			if !IsEmpty(method.Output()) {
				id := formatTypeRef(string(method.FullName() + "." + method.Output().FullName()))
				responses.WithMapOfResponseOrReferenceValuesItem("200", openapi31.ResponseOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/responses/" + id,
					},
				})
			}
			op.WithResponses(responses)

			if hasGetSupport {
				item.Get = op
				parameters = append(parameters,
					openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/encoding"}},
					openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/base64"}},
					openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/compression"}},
					openapi31.ParameterOrReference{Reference: &openapi31.Reference{Ref: "#/components/parameters/connect"}},
				)
			} else {
				item.Post = op
			}
			item.WithParameters(parameters...)
			items["/"+string(service.FullName())+"/"+string(method.Name())] = pathItemWithMethodAnnotations(item, method)
		}
	}

	return items, nil
}

func methodHasGet(method protoreflect.MethodDescriptor) bool {
	isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
	options := method.Options().(*descriptorpb.MethodOptions)
	return options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS && !isStreaming
}
