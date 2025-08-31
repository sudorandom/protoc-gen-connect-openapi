package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SpecWithFileAnnotations(opts options.Options, spec *highv3.Document, fd protoreflect.FileDescriptor) {
	if !proto.HasExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type()) {
		return
	}

	ext := proto.GetExtension(fd.Options(), goa3.E_Document.TypeDescriptor().Type())
	gnosticDocument, ok := ext.(*goa3.Document)
	if !ok {
		return
	}
	if gnosticDocument.Openapi != "" {
		spec.Info.Version = gnosticDocument.Openapi
	}

	if gnosticDocument.Info != nil {
		spec.Info.Title = gnosticDocument.Info.Title
		spec.Info.Summary = gnosticDocument.Info.Summary
		spec.Info.Description = gnosticDocument.Info.Description
		spec.Info.TermsOfService = gnosticDocument.Info.TermsOfService
		if gnosticDocument.Info.Contact != nil {
			spec.Info.Contact = &highbase.Contact{
				Name:  gnosticDocument.Info.Contact.Name,
				URL:   gnosticDocument.Info.Contact.Url,
				Email: gnosticDocument.Info.Contact.Email,
			}
		}
		if gnosticDocument.Info.License != nil {
			spec.Info.License = &highbase.License{
				Name: gnosticDocument.Info.License.Name,
				URL:  gnosticDocument.Info.License.Url,
			}
		}
		spec.Info.Version = gnosticDocument.Info.Version
	}
	spec.Servers = append(spec.Servers, toServers(gnosticDocument.Servers)...)
	spec.Security = append(spec.Security, toSecurityRequirements(gnosticDocument.Security)...)
	spec.Tags = append(spec.Tags, toTags(gnosticDocument.Tags)...)
	if exDocs := toExternalDocs(gnosticDocument.ExternalDocs); exDocs != nil {
		spec.ExternalDocs = exDocs
	}
	if gnosticDocument.SpecificationExtension != nil {
		ext := toExtensions(gnosticDocument.SpecificationExtension)
		for pair := ext.Oldest(); pair != nil; pair = pair.Next() {
			spec.Extensions.AddPairs(*pair)
		}
	}
	appendComponents(opts, spec, gnosticDocument.Components)
}
