package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/visibility"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func fileToTags(opts options.Options, fd protoreflect.FileDescriptor) []*base.Tag {
	if opts.WithoutDefaultTags {
		return nil
	}
	tags := []*base.Tag{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(service), opts.AllowedVisibilities) {
			continue
		}
		loc := fd.SourceLocations().ByDescriptor(service)
		description := util.FormatComments(loc)

		tagName := string(service.FullName())
		if opts.ShortServiceTags {
			tagName = string(service.Name())
		}
		tags = append(tags, &base.Tag{
			Name:        tagName,
			Description: description,
		})
	}
	return tags
}
