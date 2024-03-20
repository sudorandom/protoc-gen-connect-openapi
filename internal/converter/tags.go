package converter

import (
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func fileToTags(fd protoreflect.FileDescriptor) []openapi31.Tag {
	tags := []openapi31.Tag{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		loc := fd.SourceLocations().ByDescriptor(service)
		description := util.FormatComments(loc)

		tags = append(tags, openapi31.Tag{
			Name:        string(service.FullName()),
			Description: &description,
		})
	}
	return tags
}
