package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/gnostic"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/googleapi"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type annotator struct{}

func (*annotator) AnnotateMessage(schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	schema = protovalidate.SchemaWithMessageAnnotations(schema, desc)
	schema = gnostic.SchemaWithSchemaAnnotations(schema, desc)
	return schema
}

func (*annotator) AnnotateField(schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	schema = protovalidate.SchemaWithFieldAnnotations(schema, desc, onlyScalar)
	schema = gnostic.SchemaWithPropertyAnnotations(schema, desc)
	schema = googleapi.SchemaWithPropertyAnnotations(schema, desc)
	return schema
}

func (*annotator) AnnotateFieldReference(parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	parent = protovalidate.PopulateParentProperties(parent, desc)
	return parent
}
