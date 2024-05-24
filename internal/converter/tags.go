package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func fileToTags(fd protoreflect.FileDescriptor) []*base.Tag {
	tags := []*highbase.Tag{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		loc := fd.SourceLocations().ByDescriptor(service)
		description := util.FormatComments(loc)

		tags = append(tags, &base.Tag{
			Name:        string(service.FullName()),
			Description: description,
		})
	}
	return tags
}
