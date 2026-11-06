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

package converter

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/gnostic"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/googleapi"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/protovalidate"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type annotator struct{}

func (*annotator) AnnotateMessage(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	schema = protovalidate.SchemaWithMessageAnnotations(opts, schema, desc)
	return gnostic.SchemaWithSchemaAnnotations(opts, schema, desc)
}

func (*annotator) AnnotateField(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	schema = protovalidate.SchemaWithFieldAnnotations(opts, schema, desc, onlyScalar)
	schema = gnostic.SchemaWithPropertyAnnotations(opts, schema, desc)
	return googleapi.SchemaWithPropertyAnnotations(opts, schema, desc)
}

func (*annotator) AnnotateFieldReference(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	return protovalidate.PopulateParentProperties(opts, parent, desc)
}
