package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SpecWithFileAnnotations(spec *openapi31.Spec, fd protoreflect.FileDescriptor) *openapi31.Spec {
	if !proto.HasExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type()) {
		return spec
	}

	ext := proto.GetExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Document)
	if !ok {
		return spec
	}
	if opts.Openapi != "" {
		spec.Openapi = opts.Openapi
	}

	if opts.Info != nil {
		spec.Info.Title = opts.Info.Title
		spec.Info.Summary = &opts.Info.Summary
		spec.Info.Description = &opts.Info.Description
		spec.Info.TermsOfService = &opts.Info.TermsOfService
		if opts.Info.Contact != nil {
			spec.Info.Contact = &openapi31.Contact{
				Name:  &opts.Info.Contact.Name,
				URL:   &opts.Info.Contact.Url,
				Email: &opts.Info.Contact.Email,
			}
		}
		if opts.Info.License != nil {
			spec.Info.License = &openapi31.License{
				Name: opts.Info.License.Name,
				URL:  &opts.Info.License.Url,
			}
		}
		spec.Info.Version = opts.Info.Version
	}
	spec.Servers = append(spec.Servers, toServers(opts.Servers)...)
	spec.Security = append(spec.Security, toSecurityRequirements(opts.Security)...)
	for k, v := range toSecuritySchemes(opts.Components) {
		spec.Components.SecuritySchemes[k] = v
	}
	spec.Tags = append(spec.Tags, toTags(opts.Tags)...)
	if exDocs := toExternalDocs(opts.ExternalDocs); exDocs != nil {
		spec.ExternalDocs = exDocs
	}
	return spec
}
