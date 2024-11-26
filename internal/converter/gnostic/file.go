package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SpecWithFileAnnotations(spec *highv3.Document, fd protoreflect.FileDescriptor) {
	if !proto.HasExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type()) {
		return
	}

	ext := proto.GetExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Document)
	if !ok {
		return
	}
	if opts.Openapi != "" {
		spec.Info.Version = opts.Openapi
	}

	if opts.Info != nil {
		spec.Info.Title = opts.Info.Title
		spec.Info.Summary = opts.Info.Summary
		spec.Info.Description = opts.Info.Description
		spec.Info.TermsOfService = opts.Info.TermsOfService
		if opts.Info.Contact != nil {
			spec.Info.Contact = &highbase.Contact{
				Name:  opts.Info.Contact.Name,
				URL:   opts.Info.Contact.Url,
				Email: opts.Info.Contact.Email,
			}
		}
		if opts.Info.License != nil {
			spec.Info.License = &highbase.License{
				Name: opts.Info.License.Name,
				URL:  opts.Info.License.Url,
			}
		}
		spec.Info.Version = opts.Info.Version
	}
	spec.Servers = append(spec.Servers, toServers(opts.Servers)...)
	spec.Security = append(spec.Security, toSecurityRequirements(opts.Security)...)
	spec.Tags = append(spec.Tags, toTags(opts.Tags)...)
	if exDocs := toExternalDocs(opts.ExternalDocs); exDocs != nil {
		spec.ExternalDocs = exDocs
	}
	if opts.SpecificationExtension != nil {
		ext := toExtensions(opts.SpecificationExtension)
		for pair := ext.Oldest(); pair != nil; pair = pair.Next() {
			spec.Extensions.AddPairs(*pair)
		}
	}
	appendComponents(spec, opts.Components)
}
