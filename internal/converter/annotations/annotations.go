package annotations

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MessageAnnotator interface {
	AnnotateMessage(schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema
}

type FieldAnnotator interface {
	AnnotateField(schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema
}

type FieldReferenceAnnotator interface {
	// Annotate a field reference. This takes in the PARENT of the field, because with references
	// we can only annotate on the things on the parent like the list of required attributes.
	AnnotateFieldReference(parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema
}
