package converter

import (
	"log/slog"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/connectrpc"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/twirp"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/visibility"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func addPathItemsFromFile(opts options.Options, fd protoreflect.FileDescriptor, doc *v3.Document) error {
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(service), opts.AllowedVisibilities) {
			opts.Logger.Debug("Filtering service due to visibility", slog.String("service", string(service.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
			continue
		}
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)

			if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(method), opts.AllowedVisibilities) {
				opts.Logger.Debug("Filtering method due to visibility", slog.String("method", string(method.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
				continue
			}

			// No matter what, we add the schemas for the method input/output
			AddMessageSchemas(opts, method.Input(), doc)
			AddMessageSchemas(opts, method.Output(), doc)

			// Helper function to update or set path items
			addPathItem := func(path string, newItem *v3.PathItem, deferredParams []*v3.Parameter) {
				if opts.FeatureEnabled(options.FeatureGnostic) {
					newItem = gnostic.PathItemWithMethodAnnotations(opts, newItem, method)
				}
				for kv := newItem.GetOperations().First(); kv != nil; kv = kv.Next() {
					for _, dp := range deferredParams {
						for _, ep := range kv.Value().Parameters {
							if ep.Name == dp.Name && ep.In == dp.In && ep.Description == "" {
								ep.Description = dp.Description
							}
						}
					}
				}
				path = util.MakePath(opts, path)
				if existing, ok := doc.Paths.PathItems.Get(path); !ok {
					doc.Paths.PathItems.Set(path, newItem)
				} else {
					util.MergePathItems(existing, newItem)
					doc.Paths.PathItems.Set(path, existing)
				}
			}

			var isGoogleHTTP bool
			if opts.FeatureEnabled(options.FeatureGoogleAPIHTTP) {
				var result *googleapi.PathItemsResult
				result, isGoogleHTTP = googleapi.MakePathItems(opts, method)

				if result != nil {
					for pair := result.PathItems.First(); pair != nil; pair = pair.Next() {
						deferred, _ := result.DeferredParams.Get(pair.Key())
						addPathItem(pair.Key(), pair.Value(), deferred)
					}
				}
				if isGoogleHTTP {
					googleapi.AddSchemas(opts, doc)
				}
			}

			// Default to ConnectRPC/gRPC path if no google.api annotations
			if !isGoogleHTTP && opts.FeatureEnabled(options.FeatureConnectRPC) {
				pathItems := connectrpc.MakePathItems(opts, service, method)
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value(), nil)
				}
				connectrpc.AddSchemas(opts, doc, method)
			}

			if opts.FeatureEnabled(options.FeatureTwirp) {
				pathItems := twirp.MakePathItems(opts, service, method)
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value(), nil)
				}
				twirp.AddSchemas(opts, doc, method)
			}
		}
	}

	return nil
}
