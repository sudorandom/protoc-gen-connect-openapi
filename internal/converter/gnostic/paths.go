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
	kv := operations.First()
	for kv != nil {
		oper := kv.Value()
		if opts.Deprecated {
			t := true
			oper.Deprecated = &t
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
		oper.Tags = append(oper.Tags, opts.Tags...)

		if exDocs := toExternalDocs(opts.ExternalDocs); exDocs != nil {
			oper.ExternalDocs = exDocs
		}

		if opts.OperationId != "" {
			oper.OperationId = opts.OperationId
		}

		kv = kv.Next()
	}
	return item
}
