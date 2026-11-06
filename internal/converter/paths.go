package converter

import (
	"log/slog"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
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

			// A visible method whose request or response type is hidden would
			// reference a schema that's never emitted, so the method is hidden
			// too.
			if hidden := hiddenMessage(opts, method.Input(), method.Output()); hidden != nil {
				opts.Logger.Warn("Filtering method because its message is hidden by visibility",
					slog.String("method", string(method.FullName())), slog.String("message", string(hidden.FullName())))
				continue
			}

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
					util.MergePathItems(existing, newItem)
					doc.Paths.PathItems.Set(path, existing)
				}
			}

			var isGoogleHTTP bool
			if opts.FeatureEnabled(options.FeatureGoogleAPIHTTP) {
				var pathItems *orderedmap.Map[string, *v3.PathItem]
				pathItems, isGoogleHTTP = googleapi.MakePathItems(opts, method)
				for pair := pathItems.First(); pair != nil; pair = pair.Next() {
					addPathItem(pair.Key(), pair.Value())
				}
				if isGoogleHTTP {
					googleapi.AddSchemas(opts, doc)
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

func hiddenMessage(opts options.Options, messages ...protoreflect.MessageDescriptor) protoreflect.MessageDescriptor {
	for _, md := range messages {
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(md), opts.AllowedVisibilities) {
			return md
		}
	}
	return nil
}
