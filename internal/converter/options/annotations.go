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

package options

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MessageAnnotator interface {
	AnnotateMessage(opts Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema
}

type FieldAnnotator interface {
	AnnotateField(opts Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema
}

type FieldReferenceAnnotator interface {
	// Annotate a field reference. This takes in the PARENT of the field, because with references
	// we can only annotate on the things on the parent like the list of required attributes.
	AnnotateFieldReference(opts Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema
}
