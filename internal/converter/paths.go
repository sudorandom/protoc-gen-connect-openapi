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
		loc := fd.SourceLocations().ByDescriptor(service)
		description := formatComments(loc)

		methods := service.Methods()
		for j := 0; j < services.Len(); j++ {
			method := methods.Get(j)
			op := &openapi31.Operation{}
			op.WithTags(string(service.FullName()))
			loc := fd.SourceLocations().ByDescriptor(method)
			op.WithDescription(formatComments(loc))

			// Request Body
			item := openapi31.PathItem{}
			op.WithRequestBody(openapi31.RequestBodyOrReference{
				Reference: &openapi31.Reference{
					Ref: "#/components/requestBodies/" + formatTypeRef(string(method.Input().FullName())),
				},
			})

			// Responses
			op.WithResponses(openapi31.Responses{
				Default: &openapi31.ResponseOrReference{
					Response: &openapi31.Response{
						Content: map[string]openapi31.MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"$ref": "#/components/responses/connect.error",
								},
							},
						},
					},
				},
				MapOfResponseOrReferenceValues: map[string]openapi31.ResponseOrReference{
					"200": {
						Response: &openapi31.Response{
							Description: description,
							Content: map[string]openapi31.MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"$ref": "#/components/responses/" + formatTypeRef(string(method.Output().FullName())),
									},
								},
							},
						},
					},
				},
			})

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
