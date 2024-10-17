package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func addPathItemsFromFile(opts options.Options, fd protoreflect.FileDescriptor, paths *v3.Paths) error {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				pathItems := googleapi.MakePathItems(opts, method)
				
				// Helper function to update or set path items
				updatePathItem := func(path string, newItem *v3.PathItem) {
					if existing, ok := paths.PathItems.Get(path); !ok {
						paths.PathItems.Set(path, newItem)
					} else {
						if opts.MergeBase {
							mergePathItems(existing, newItem)
						} else {
							updateOperations(existing, newItem)
						}
						paths.PathItems.Set(path, existing)
					}
				}

				// Update path items from google.api annotations
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					updatePathItem(pair.Key(), pair.Value())
				}

				// Default to ConnectRPC/gRPC path if no google.api annotations
				if pathItems == nil || pathItems.Len() == 0 {
					path := "/" + string(service.FullName()) + "/" + string(method.Name())
					updatePathItem(path, methodToPathItem(opts, method))
				}
			}
		}

	return nil
}

// Helper function to update operations
func updateOperations(existing, newItem *v3.PathItem) {
	operations := []struct {
		existingOp **v3.Operation
		newOp      *v3.Operation
	}{
		{&existing.Get, newItem.Get},
		{&existing.Post, newItem.Post},
		{&existing.Delete, newItem.Delete},
		{&existing.Put, newItem.Put},
		{&existing.Patch, newItem.Patch},
	}

	for _, op := range operations {
		if op.newOp != nil {
			*op.existingOp = op.newOp
		}
	}
}

func mergePathItems(existing, new *v3.PathItem) {
	// Merge operations
	if new.Get != nil {
		mergeOperation(existing, new, &existing.Get, new.Get)
	}
	if new.Post != nil {
		mergeOperation(existing, new, &existing.Post, new.Post)
	}
	if new.Put != nil {
		mergeOperation(existing, new, &existing.Put, new.Put)
	}
	if new.Delete != nil {
		mergeOperation(existing, new, &existing.Delete, new.Delete)
	}
	if new.Options != nil {
		mergeOperation(existing, new, &existing.Options, new.Options)
	}
	if new.Head != nil {
		mergeOperation(existing, new, &existing.Head, new.Head)
	}
	if new.Patch != nil {
		mergeOperation(existing, new, &existing.Patch, new.Patch)
	}
	if new.Trace != nil {
		mergeOperation(existing, new, &existing.Trace, new.Trace)
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

func mergeOperation(existingItem, newItem *v3.PathItem, existing **v3.Operation, new *v3.Operation) {
	if *existing == nil {
		*existing = new
	} else {
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

		// Merge extensions
		for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
			if _, ok := (*existing).Extensions.Get(pair.Key()); !ok {
				(*existing).Extensions.Set(pair.Key(), pair.Value())
			}
		}
	}
}

func mergeResponses(existing, new *v3.Responses) {
	if existing == nil {
		return
	}
	if new == nil {
		return
	}

	// Merge response codes
	for pair := new.Codes.First(); pair != nil; pair = pair.Next() {
		if _, ok := existing.Codes.Get(pair.Key()); !ok {
			existing.Codes.Set(pair.Key(), pair.Value())
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
	for pair := new.Content.First(); pair != nil; pair = pair.Next() {
		contentType := pair.Key()
		mediaType := pair.Value()
		if _, ok := existing.Content.Get(contentType); !ok {
			existing.Content.Set(contentType, mediaType)
		}
	}
	for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
		if _, ok := existing.Extensions.Get(pair.Key()); !ok {
			existing.Extensions.Set(pair.Key(), pair.Value())
		}
	}
}

func methodToOperaton(opts options.Options, method protoreflect.MethodDescriptor, returnGet bool) *v3.Operation {
	fd := method.ParentFile()
	service := method.Parent().(protoreflect.ServiceDescriptor)
	loc := fd.SourceLocations().ByDescriptor(method)
	op := &v3.Operation{
		Summary:     string(method.Name()),
		OperationId: string(method.FullName()),
		Deprecated:  util.IsMethodDeprecated(method),
		Tags:        []string{string(service.FullName())},
		Description: util.FormatComments(loc),
	}

	isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
	if isStreaming && !opts.WithStreaming {
		return nil
	}

	// Responses
	codeMap := orderedmap.New[string, *v3.Response]()
	outputId := util.FormatTypeRef(string(method.Output().FullName()))
	codeMap.Set("200", &v3.Response{
		Description: "Success",
		Content: util.MakeMediaTypes(
			opts,
			base.CreateSchemaProxyRef("#/components/schemas/"+outputId),
			false,
			isStreaming,
		),
	})
	op.Responses = &v3.Responses{
		Codes: codeMap,
		Default: &v3.Response{
			Description: "Error",
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/connect.error"),
				false,
				isStreaming,
			),
		},
	}

	op.Parameters = append(op.Parameters,
		&v3.Parameter{
			Name:     "Connect-Protocol-Version",
			In:       "header",
			Required: util.BoolPtr(true),
			Schema:   base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
		},
		&v3.Parameter{
			Name:   "Connect-Timeout-Ms",
			In:     "header",
			Schema: base.CreateSchemaProxyRef("#/components/schemas/connect-timeout-header"),
		},
	)

	// Request parameters
	inputId := util.FormatTypeRef(string(method.Input().FullName()))
	if returnGet {
		op.OperationId = op.OperationId + ".get"
		op.Parameters = append(op.Parameters,
			&v3.Parameter{
				Name: "message",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/"+util.FormatTypeRef(inputId)),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "encoding",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/encoding"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "base64",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/base64"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "compression",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/compression"),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name: "connect",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
					true,
					isStreaming),
			},
		)
	} else {
		op.RequestBody = &v3.RequestBody{
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/"+inputId),
				true,
				isStreaming,
			),
			Required: util.BoolPtr(true),
		}
	}

	return op
}

func methodToPathItem(opts options.Options, method protoreflect.MethodDescriptor) *v3.PathItem {
	hasGetSupport := methodHasGet(opts, method)
	item := &v3.PathItem{}
	if hasGetSupport {
		item.Get = methodToOperaton(opts, method, true)
	}
	item.Post = methodToOperaton(opts, method, false)
	item = gnostic.PathItemWithMethodAnnotations(item, method)

	return item
}

func methodHasGet(opts options.Options, method protoreflect.MethodDescriptor) bool {
	if !opts.AllowGET {
		return false
	}

	if method.IsStreamingClient() || method.IsStreamingServer() {
		return false
	}

	options := method.Options().(*descriptorpb.MethodOptions)
	return options.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS
}
