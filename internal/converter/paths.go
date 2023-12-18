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

			// Request Body
			item := openapi31.PathItem{}
			if !IsEmpty(method.Input()) {
				op.WithRequestBody(openapi31.RequestBodyOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/requestBodies/" + formatTypeRef(string(method.Input().FullName())),
					},
				})
			}

			// Responses
			responses := openapi31.Responses{
				Default: &openapi31.ResponseOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/responses/connect.error",
					},
				},
			}
			if !IsEmpty(method.Input()) {
				responses.WithMapOfResponseOrReferenceValuesItem("200", openapi31.ResponseOrReference{
					Reference: &openapi31.Reference{
						Ref: "#/components/responses/" + formatTypeRef(string(method.Output().FullName())),
					},
				})
			}
			op.WithResponses(responses)

			options := method.Options().(*descriptorpb.MethodOptions)
			if options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS {
				item.Get = op
			} else {
				item.Post = op
			}
			items["/"+string(service.FullName())+"/"+string(method.Name())] = item
		}
	}

	return items, nil
}
