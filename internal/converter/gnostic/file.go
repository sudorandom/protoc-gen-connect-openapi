// Copyright 2023-2026 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnostic

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/go-yaml"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
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
	// gnosticDocument.Openapi is the OpenAPI spec version, which this plugin
	// controls; it isn't the API's version, so it's not copied anywhere.

	// Only fields the annotation sets replace the defaults: proto3 strings
	// can't tell an empty value from an unset one.
	if info := gnosticDocument.Info; info != nil {
		setIfNotEmpty(&spec.Info.Title, info.Title)
		setIfNotEmpty(&spec.Info.Summary, info.Summary)
		setIfNotEmpty(&spec.Info.Description, info.Description)
		setIfNotEmpty(&spec.Info.TermsOfService, info.TermsOfService)
		setIfNotEmpty(&spec.Info.Version, info.Version)
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
	if gnosticDocument.Paths != nil {
		if spec.Paths == nil {
			spec.Paths = &highv3.Paths{}
		}
		if spec.Paths.PathItems == nil {
			spec.Paths.PathItems = orderedmap.New[string, *highv3.PathItem]()
		}
		if spec.Paths.Extensions == nil {
			spec.Paths.Extensions = orderedmap.New[string, *yaml.Node]()
		}
		for _, item := range gnosticDocument.Paths.GetPath() {
			if existing, ok := spec.Paths.PathItems.Get(item.Name); ok {
				util.MergePathItems(existing, toPathItem(opts, item.Value))
			} else {
				spec.Paths.PathItems.Set(item.Name, toPathItem(opts, item.Value))
			}
		}
		if gnosticDocument.Paths.SpecificationExtension != nil {
			ext := toExtensions(gnosticDocument.Paths.SpecificationExtension)
			for pair := ext.Oldest(); pair != nil; pair = pair.Next() {
				spec.Paths.Extensions.AddPairs(*pair)
			}
		}
	}
	appendComponents(opts, spec, gnosticDocument.Components)
}

func setIfNotEmpty(target *string, value string) {
	if value != "" {
		*target = value
	}
}
