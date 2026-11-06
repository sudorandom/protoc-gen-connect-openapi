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
	"strings"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	goa3 "github.com/google/gnostic/openapiv3"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
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
		method, oper := kv.Key(), kv.Value()
		if gnosticOperation.Deprecated {
			t := true
			oper.Deprecated = &t
		}

		oper.Parameters = util.MergeParameters(oper.Parameters, toParameters(opts, gnosticOperation.Parameters))

		// GET and HEAD carry no body; for allow-get the request goes in the
		// query string, so a request body only belongs on the other methods.
		if gnosticOperation.RequestBody != nil && method != "get" && method != "head" {
			oper.RequestBody = toRequestBody(opts, gnosticOperation.RequestBody)
		}

		if gnosticOperation.Responses != nil {
			responses := toResponses(opts, gnosticOperation.Responses)
			for pair := responses.Codes.First(); pair != nil; pair = pair.Next() {
				if pair.Key() == "default" {
					// Codes can't hold "default" alongside Responses.Default
					// without emitting the key twice.
					oper.Responses.Default = pair.Value()
					continue
				}
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

		// One method can produce several operations (allow-get's ".get",
		// numbered additional bindings) whose IDs extend the default one. Keep
		// those suffixes so the custom ID doesn't collide with itself.
		if gnosticOperation.OperationId != "" {
			if suffix, ok := strings.CutPrefix(oper.OperationId, util.DefaultOperationID(opts, md)); ok {
				oper.OperationId = gnosticOperation.OperationId + suffix
			} else {
				oper.OperationId = gnosticOperation.OperationId
			}
		}

		if gnosticOperation.SpecificationExtension != nil {
			oper.Extensions = toExtensions(gnosticOperation.GetSpecificationExtension())
		}
	}
	return item
}
