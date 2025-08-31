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
	schema = protovalidate.SchemaWithMessageAnnotations(opts, schema, desc)
	schema = gnostic.SchemaWithSchemaAnnotations(opts, schema, desc)
	return schema
}

func (*annotator) AnnotateField(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	schema = protovalidate.SchemaWithFieldAnnotations(opts, schema, desc, onlyScalar)
	schema = gnostic.SchemaWithPropertyAnnotations(opts, schema, desc)
	schema = googleapi.SchemaWithPropertyAnnotations(opts, schema, desc)
	return schema
}

func (*annotator) AnnotateFieldReference(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	parent = protovalidate.PopulateParentProperties(opts, parent, desc)
	return parent
}
