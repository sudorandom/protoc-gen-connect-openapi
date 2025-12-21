package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func PathItemWithMethodAnnotations(opts options.Options, item *v3.PathItem, md protoreflect.MethodDescriptor) *v3.PathItem {
	if !proto.HasExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type()) {
		return item
	}

	ext := proto.GetExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type())
	gnosticOperation, ok := ext.(*goa3.Operation)
	if !ok {
		return item
	}
	operations := item.GetOperations()
	for kv := operations.First(); kv != nil; kv = kv.Next() {
		oper := kv.Value()
		if gnosticOperation.Deprecated {
			t := true
			oper.Deprecated = &t
		}

		for _, param := range gnosticOperation.Parameters {
			item.Parameters = append(item.Parameters, toParameter(opts, param))
		}

		if gnosticOperation.RequestBody != nil {
			oper.RequestBody = toRequestBody(opts, gnosticOperation.RequestBody.GetRequestBody())
		}

		if gnosticOperation.Responses != nil {
			responses := toResponses(opts, gnosticOperation.Responses)
			for pair := responses.Codes.First(); pair != nil; pair = pair.Next() {
				oper.Responses.Codes.Set(pair.Key(), pair.Value())
			}
			if responses.Default != nil {
				oper.Responses.Default = responses.Default
			}
			for pair := responses.Extensions.First(); pair != nil; pair = pair.Next() {
				oper.Responses.Extensions.Set(pair.Key(), pair.Value())
			}
		}

		if gnosticOperation.Callbacks != nil {
			oper.Callbacks = toCallbacks(opts, gnosticOperation.Callbacks)
		}

		if gnosticOperation.Security != nil {
			oper.Security = toSecurityRequirements(gnosticOperation.Security)
		}
		oper.Servers = toServers(gnosticOperation.Servers)

		if gnosticOperation.Summary != "" {
			oper.Summary = gnosticOperation.Summary
		}
		if gnosticOperation.Description != "" {
			oper.Description = gnosticOperation.Description
		}
		oper.Tags = append(gnosticOperation.Tags, oper.Tags...)

		if exDocs := toExternalDocs(gnosticOperation.ExternalDocs); exDocs != nil {
			oper.ExternalDocs = exDocs
		}

		if gnosticOperation.OperationId != "" {
			oper.OperationId = gnosticOperation.OperationId
		}

		if gnosticOperation.SpecificationExtension != nil {
			oper.Extensions = toExtensions(gnosticOperation.GetSpecificationExtension())
		}
	}
	return item
}
