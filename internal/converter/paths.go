package converter

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/connectrpc"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/twirp"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func addPathItemsFromFile(opts options.Options, fd protoreflect.FileDescriptor, doc *v3.Document) error {
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)

			// No matter what, we add the schemas for the method input/output
			AddMessageSchemas(opts, method.Input(), doc)
			AddMessageSchemas(opts, method.Output(), doc)

			// Helper function to update or set path items
			addPathItem := func(path string, newItem *v3.PathItem) {
				if opts.FeatureEnabled(options.FeatureGnostic) {
					newItem = gnostic.PathItemWithMethodAnnotations(opts, newItem, method)
				}
				path = util.MakePath(opts, path)
				if existing, ok := doc.Paths.PathItems.Get(path); !ok {
					doc.Paths.PathItems.Set(path, newItem)
				} else {
					mergePathItems(existing, newItem)
					doc.Paths.PathItems.Set(path, existing)
				}
			}

			var isGoogleHTTP bool
			if opts.FeatureEnabled(options.FeatureGoogleAPIHTTP) {
				var pathItems *orderedmap.Map[string, *v3.PathItem]
				pathItems, isGoogleHTTP = googleapi.MakePathItems(opts, method)

				// Update path items from google.api annotations
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value())
				}
			}

			// Default to ConnectRPC/gRPC path if no google.api annotations
			if !isGoogleHTTP && opts.FeatureEnabled(options.FeatureConnectRPC) {
				pathItems := connectrpc.MakePathItems(opts, service, method)
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value())
				}
				connectrpc.AddSchemas(opts, doc, method)
			}

			if opts.FeatureEnabled(options.FeatureTwirp) {
				pathItems := twirp.MakePathItems(opts, service, method)
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value())
				}
				twirp.AddSchemas(opts, doc, method)
			}
		}
	}

	return nil
}

func mergePathItems(existing, new *v3.PathItem) {
	// Merge operations
	operations := []struct {
		existingOp **v3.Operation
		newOp      *v3.Operation
	}{
		{&existing.Get, new.Get},
		{&existing.Post, new.Post},
		{&existing.Put, new.Put},
		{&existing.Delete, new.Delete},
		{&existing.Options, new.Options},
		{&existing.Head, new.Head},
		{&existing.Patch, new.Patch},
		{&existing.Trace, new.Trace},
	}

	for _, op := range operations {
		if op.newOp != nil {
			mergeOperation(op.existingOp, op.newOp)
		}
	}

	// Merge other fields
	if new.Summary != "" {
		existing.Summary = new.Summary
	}
	if new.Description != "" {
		existing.Description = new.Description
	}
	existing.Servers = append(existing.Servers, new.Servers...)
	existing.Parameters = append(existing.Parameters, new.Parameters...)

	// Merge extensions
	for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
		if _, ok := existing.Extensions.Get(pair.Key()); !ok {
			existing.Extensions.Set(pair.Key(), pair.Value())
		}
	}
}

func mergeOperation(existing **v3.Operation, new *v3.Operation) {
	if *existing == nil {
		*existing = new
		return
	}
	// Merge operation fields
	if new.Summary != "" {
		(*existing).Summary = new.Summary
	}
	if new.Description != "" {
		(*existing).Description = new.Description
	}
	(*existing).Tags = append((*existing).Tags, new.Tags...)
	(*existing).Parameters = append((*existing).Parameters, new.Parameters...)
	if new.RequestBody != nil {
		(*existing).RequestBody = new.RequestBody
	}
	if new.Responses != nil {
		mergeResponses((*existing).Responses, new.Responses)
	}
	if new.Deprecated != nil {
		(*existing).Deprecated = new.Deprecated
	}

	// Add support for additional Operation fields
	if new.Callbacks != nil {
		if (*existing).Callbacks == nil {
			(*existing).Callbacks = orderedmap.New[string, *v3.Callback]()
		}
		for pair := new.Callbacks.First(); pair != nil; pair = pair.Next() {
			if _, ok := (*existing).Callbacks.Get(pair.Key()); !ok {
				(*existing).Callbacks.Set(pair.Key(), pair.Value())
			}
		}
	}

	if new.Security != nil {
		(*existing).Security = append((*existing).Security, new.Security...)
	}

	if new.Servers != nil {
		(*existing).Servers = append((*existing).Servers, new.Servers...)
	}

	if new.ExternalDocs != nil {
		(*existing).ExternalDocs = new.ExternalDocs
	}

	// Merge extensions
	for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
		if _, ok := (*existing).Extensions.Get(pair.Key()); !ok {
			(*existing).Extensions.Set(pair.Key(), pair.Value())
		}
	}
}

func mergeResponses(existing, new *v3.Responses) {
	if existing == nil || new == nil {
		return
	}

	// Merge response codes
	for pair := new.Codes.First(); pair != nil; pair = pair.Next() {
		code := pair.Key()
		if existingResponse, ok := existing.Codes.Get(code); !ok {
			existing.Codes.Set(code, pair.Value())
		} else {
			mergeResponse(existingResponse, pair.Value())
		}
	}

	// Merge default response
	if new.Default != nil {
		if existing.Default == nil {
			existing.Default = new.Default
		} else {
			mergeResponse(existing.Default, new.Default)
		}
	}
}

func mergeResponse(existing, new *v3.Response) {
	if new.Description != "" {
		existing.Description = new.Description
	}

	// Merge Content
	for pair := new.Content.First(); pair != nil; pair = pair.Next() {
		contentType := pair.Key()
		mediaType := pair.Value()
		if _, ok := existing.Content.Get(contentType); !ok {
			existing.Content.Set(contentType, mediaType)
		}
	}

	// Merge Headers
	if new.Headers != nil {
		if existing.Headers == nil {
			existing.Headers = orderedmap.New[string, *v3.Header]()
		}
		for pair := new.Headers.First(); pair != nil; pair = pair.Next() {
			if _, ok := existing.Headers.Get(pair.Key()); !ok {
				existing.Headers.Set(pair.Key(), pair.Value())
			}
		}
	}

	// Merge Links
	if new.Links != nil {
		if existing.Links == nil {
			existing.Links = orderedmap.New[string, *v3.Link]()
		}
		for pair := new.Links.First(); pair != nil; pair = pair.Next() {
			if _, ok := existing.Links.Get(pair.Key()); !ok {
				existing.Links.Set(pair.Key(), pair.Value())
			}
		}
	}

	// Merge Extensions
	for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
		if _, ok := existing.Extensions.Get(pair.Key()); !ok {
			existing.Extensions.Set(pair.Key(), pair.Value())
		}
	}
}
