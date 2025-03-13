package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func PathItemWithMethodAnnotations(item *v3.PathItem, md protoreflect.MethodDescriptor) *v3.PathItem {
	if !proto.HasExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type()) {
		return item
	}

	ext := proto.GetExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Operation)
	if !ok {
		return item
	}
	operations := item.GetOperations()
	for kv := operations.First(); kv != nil; kv = kv.Next() {
		oper := kv.Value()
		if opts.Deprecated {
			t := true
			oper.Deprecated = &t
		}

		for _, param := range opts.Parameters {
			item.Parameters = append(item.Parameters, toParameter(param))
		}

		if opts.RequestBody != nil {
			oper.RequestBody = toRequestBody(opts.RequestBody.GetRequestBody())
		}

		if opts.Responses != nil {
			responses := toResponses(opts.Responses)
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

		if opts.Callbacks != nil {
			oper.Callbacks = toCallbacks(opts.Callbacks)
		}

		if security := toSecurityRequirements(opts.Security); len(security) > 0 {
			oper.Security = security
		}
		oper.Servers = toServers(opts.Servers)

		if opts.Summary != "" {
			oper.Summary = opts.Summary
		}
		if opts.Description != "" {
			oper.Description = opts.Description
		}
		oper.Tags = append(opts.Tags, oper.Tags...)

		if exDocs := toExternalDocs(opts.ExternalDocs); exDocs != nil {
			oper.ExternalDocs = exDocs
		}

		if opts.OperationId != "" {
			oper.OperationId = opts.OperationId
		}

		if opts.SpecificationExtension != nil {
			oper.Extensions = toExtensions(opts.GetSpecificationExtension())
		}
	}
	return item
}
