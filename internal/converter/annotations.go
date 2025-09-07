package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type annotator struct{}

func (*annotator) AnnotateMessage(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	if opts.FeatureEnabled(options.FeatureProtovalidate) {
		schema = protovalidate.SchemaWithMessageAnnotations(opts, schema, desc)
	}
	if opts.FeatureEnabled(options.FeatureGnostic) {
		schema = gnostic.SchemaWithSchemaAnnotations(opts, schema, desc)
	}
	return schema
}

func (*annotator) AnnotateField(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	if opts.FeatureEnabled(options.FeatureProtovalidate) {
		schema = protovalidate.SchemaWithFieldAnnotations(opts, schema, desc, onlyScalar)
	}
	if opts.FeatureEnabled(options.FeatureGnostic) {
		schema = gnostic.SchemaWithPropertyAnnotations(opts, schema, desc)
	}
	if opts.FeatureEnabled(options.FeatureGoogleAPIHTTP) {
		schema = googleapi.SchemaWithPropertyAnnotations(opts, schema, desc)
	}
	return schema
}

func (*annotator) AnnotateFieldReference(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	if opts.FeatureEnabled(options.FeatureProtovalidate) {
		parent = protovalidate.PopulateParentProperties(opts, parent, desc)
	}
	return parent
}
