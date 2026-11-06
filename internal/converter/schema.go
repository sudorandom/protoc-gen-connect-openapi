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
	"log/slog"
	"strconv"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/reflect/protoreflect"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/schema"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/visibility"
)

func AddMessageSchemas(opts options.Options, md protoreflect.MessageDescriptor, doc *v3.Document) {
	if md == nil {
		return
	}
	// Synthetic map-entry messages never appear in JSON: map fields are
	// rendered inline as objects with additionalProperties, so nothing can
	// reference an entry schema.
	if md.IsMapEntry() {
		return
	}
	if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(md), opts.AllowedVisibilities) {
		opts.Logger.Debug("Filtering message due to visibility", slog.String("message", string(md.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
		return
	}
	if _, ok := doc.Components.Schemas.Get(string(md.FullName())); ok {
		return
	}
	name, schema := schema.MessageToSchema(opts, md)
	if schema != nil {
		doc.Components.Schemas.Set(name, base.CreateSchemaProxy(schema))
	}

	// Messages can have fields. Fields hidden by visibility aren't in the
	// schema, so their types aren't pulled in either.
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(field), opts.AllowedVisibilities) {
			continue
		}
		AddFieldToSchema(opts, field, doc)
	}

	// Messages can have enums
	enums := md.Enums()
	for i := 0; i < enums.Len(); i++ {
		AddEnumToSchema(opts, enums.Get(i), doc)
	}

	// Messages can have messages
	messages := md.Messages()
	for i := 0; i < messages.Len(); i++ {
		AddMessageSchemas(opts, messages.Get(i), doc)
	}
}

func AddFieldToSchema(opts options.Options, fd protoreflect.FieldDescriptor, doc *v3.Document) {
	if fd == nil {
		return
	}
	AddEnumToSchema(opts, fd.Enum(), doc)
	AddMessageSchemas(opts, fd.Message(), doc)
	AddFieldToSchema(opts, fd.MapKey(), doc)
	AddFieldToSchema(opts, fd.MapValue(), doc)
}

func AddEnumToSchema(opts options.Options, ed protoreflect.EnumDescriptor, doc *v3.Document) {
	if ed == nil {
		return
	}
	if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(ed), opts.AllowedVisibilities) {
		opts.Logger.Debug("Filtering enum due to visibility", slog.String("enum", string(ed.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
		return
	}
	if _, ok := doc.Components.Schemas.Get(string(ed.FullName())); ok {
		return
	}
	name, schema := enumToSchema(opts, ed)
	if schema != nil {
		doc.Components.Schemas.Set(name, base.CreateSchemaProxy(schema))
	}
}

func enumToSchema(opts options.Options, tt protoreflect.EnumDescriptor) (string, *base.Schema) {
	opts.Logger.Debug("enumToSchema", slog.Any("descriptor", tt.FullName()))
	// NullValue is an enum in proto, but protojson writes it as JSON null.
	if tt.FullName() == "google.protobuf.NullValue" {
		return string(tt.FullName()), &base.Schema{
			Title:       string(tt.Name()),
			Description: "The JSON `null` value.",
			Type:        []string{"null"},
		}
	}
	children := []*yaml.Node{}
	values := tt.Values()
	for i := 0; i < values.Len(); i++ {
		value := values.Get(i)
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(value), opts.AllowedVisibilities) {
			opts.Logger.Debug("Filtering enum value due to visibility", slog.String("enum_value", string(value.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
			continue // Skip this enum value
		}
		children = append(children, utils.CreateStringNode(string(value.Name())))
		if opts.IncludeNumberEnumValues {
			children = append(children, utils.CreateIntNode(strconv.FormatInt(int64(value.Number()), 10)))
		}
	}

	title := string(tt.Name())
	if opts.FullyQualifiedMessageNames {
		title = string(tt.FullName())
	}
	types := []string{"string"}
	if opts.IncludeNumberEnumValues {
		types = append(types, "number")
	}
	s := &base.Schema{
		Title:       title,
		Description: util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)),
		Type:        types,
		Enum:        children,
	}
	return string(tt.FullName()), s
}
