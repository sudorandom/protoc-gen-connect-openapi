package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/swaggest/openapi-go/openapi31"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func PathItemWithMethodAnnotations(item openapi31.PathItem, md protoreflect.MethodDescriptor) openapi31.PathItem {
	if !proto.HasExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type()) {
		return item
	}

	ext := proto.GetExtension(md.Options(), goa3.E_Operation.TypeDescriptor().Type())
	opts, ok := ext.(*goa3.Operation)
	if !ok {
		return item
	}
	for _, oper := range getAllOperations(item) {
		if opts.Deprecated {
			t := true
			oper.Deprecated = &t
		}

		if security := toSecurityRequirements(opts.Security); len(security) > 0 {
			oper.Security = security
		}
		oper.Servers = toServers(opts.Servers)

		if opts.Summary != "" {
			oper.Summary = &opts.Summary
		}
		if opts.Description != "" {
			oper.Description = &opts.Description
		}
		oper.Tags = append(oper.Tags, opts.Tags...)

		if exDocs := toExternalDocs(opts.ExternalDocs); exDocs != nil {
			oper.ExternalDocs = exDocs
		}

		if opts.OperationId != "" {
			oper.ID = &opts.OperationId
		}
	}
	return item
}

func getAllOperations(item openapi31.PathItem) []*openapi31.Operation {
	operations := []*openapi31.Operation{}
	if item.Get != nil {
		operations = append(operations, item.Get)
	}
	if item.Post != nil {
		operations = append(operations, item.Post)
	}
	if item.Put != nil {
		operations = append(operations, item.Put)
	}
	if item.Delete != nil {
		operations = append(operations, item.Delete)
	}
	if item.Head != nil {
		operations = append(operations, item.Head)
	}
	if item.Patch != nil {
		operations = append(operations, item.Patch)
	}
	if item.Options != nil {
		operations = append(operations, item.Options)
	}
	if item.Trace != nil {
		operations = append(operations, item.Trace)
	}
	return operations
}
