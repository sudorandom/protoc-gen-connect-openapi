package converter

import (
	"log/slog"
	"strconv"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/utils"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/schema"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
)

func AddMessageSchemas(opts options.Options, md protoreflect.MessageDescriptor, doc *v3.Document) {
	if md == nil {
		return
	}
	if _, ok := doc.Components.Schemas.Get(string(md.FullName())); ok {
		return
	}
	name, schema := schema.MessageToSchema(opts, md)
	if schema != nil {
		doc.Components.Schemas.Set(name, base.CreateSchemaProxy(schema))
	}

	// Messages can have fields
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		AddFieldToSchema(opts, fields.Get(i), doc)
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
	children := []*yaml.Node{}
	values := tt.Values()
	for i := 0; i < values.Len(); i++ {
		value := values.Get(i)
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
