package protovalidate

import (
	"strconv"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/bufbuild/protovalidate-go/resolve"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

func SchemaWithMessageAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	constraints := resolve.MessageConstraints(desc)
	if constraints == nil || constraints.GetDisabled() {
		return schema
	}
	updateWithCEL(schema, constraints.GetCel())
	return schema
}

func SchemaWithFieldAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	constraints := resolve.FieldConstraints(desc)
	if constraints == nil {
		return schema
	}
	updateWithCEL(schema, constraints.GetCel())
	if constraints.Required != nil && *constraints.Required {
		parent := schema.ParentProxy.Schema()
		if parent != nil {
			parent.Required = util.AppendStringDedupe(parent.Required, util.MakeFieldName(opts, desc))
		}
	}
	updateSchemaWithFieldConstraints(schema, constraints, onlyScalar)
	return schema
}

func PopulateParentProperties(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	if parent == nil {
		return parent
	}
	constraints := resolve.FieldConstraints(desc)
	if constraints == nil {
		return parent
	}
	if constraints.Required != nil && *constraints.Required {
		parent.Required = util.AppendStringDedupe(parent.Required, util.MakeFieldName(opts, desc))
	}
	return parent
}

//gocyclo:ignore
func updateSchemaWithFieldConstraints(schema *base.Schema, constraints *validate.FieldConstraints, onlyScalar bool) {
	if constraints == nil {
		return
	}
	switch t := constraints.Type.(type) {
	case *validate.FieldConstraints_Float:
		updateSchemaFloat(schema, t.Float)
	case *validate.FieldConstraints_Double:
		updateSchemaDouble(schema, t.Double)
	case *validate.FieldConstraints_Int32:
		updateSchemaInt32(schema, t.Int32)
	case *validate.FieldConstraints_Int64:
		updateSchemaInt64(schema, t.Int64)
	case *validate.FieldConstraints_Uint32:
		updateSchemaUint32(schema, t.Uint32)
	case *validate.FieldConstraints_Uint64:
		updateSchemaUint64(schema, t.Uint64)
	case *validate.FieldConstraints_Sint32:
		updateSchemaSint32(schema, t.Sint32)
	case *validate.FieldConstraints_Sint64:
		updateSchemaSint64(schema, t.Sint64)
	case *validate.FieldConstraints_Fixed32:
		updateSchemaFixed32(schema, t.Fixed32)
	case *validate.FieldConstraints_Fixed64:
		updateSchemaFixed64(schema, t.Fixed64)
	case *validate.FieldConstraints_Sfixed32:
		updateSchemaSfixed32(schema, t.Sfixed32)
	case *validate.FieldConstraints_Sfixed64:
		updateSchemaSfixed64(schema, t.Sfixed64)
	case *validate.FieldConstraints_Bool:
		updateSchemaBool(schema, t.Bool)
	case *validate.FieldConstraints_String_:
		updateSchemaString(schema, t.String_)
	case *validate.FieldConstraints_Bytes:
		updateSchemaBytes(schema, t.Bytes)
	case *validate.FieldConstraints_Enum:
		updateSchemaEnum(schema, t.Enum)
	case *validate.FieldConstraints_Any:
		updateSchemaAny(schema, t.Any)
	case *validate.FieldConstraints_Duration:
		updateSchemaDuration(schema, t.Duration)
	case *validate.FieldConstraints_Timestamp:
		updateSchemaTimestamp(schema, t.Timestamp)
	}

	if !onlyScalar {
		switch t := constraints.Type.(type) {
		case *validate.FieldConstraints_Repeated:
			updateSchemaRepeated(schema, t.Repeated)
		case *validate.FieldConstraints_Map:
			updateSchemaMap(schema, t.Map)
		}
	}
}

func updateWithCEL(schema *base.Schema, constraints []*validate.Constraint) {
	if len(constraints) == 0 {
		return
	}
	b := strings.Builder{}
	if schema.Description != "" {
		b.WriteString(schema.Description)
		b.WriteByte('\n')
	}
	for _, cel := range constraints {
		if cel.Message != nil {
			b.WriteString(*cel.Message)
			b.WriteString(":\n```\n")
		}
		if cel.Expression != nil {
			b.WriteString(*cel.Expression)
			b.WriteString("\n```\n\n")
		}
	}
	s := b.String()
	schema.Description = s
}

func updateSchemaFloat(schema *base.Schema, constraint *validate.FloatRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatFloat(float64(*constraint.Const), 'f', -1, 32))
		switch tt := constraint.LessThan.(type) {
		case *validate.FloatRules_Lt:
			v := float64(tt.Lt)
			schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
		case *validate.FloatRules_Lte:
			v := float64(tt.Lte)
			schema.Maximum = &v
		}
		switch tt := constraint.GreaterThan.(type) {
		case *validate.FloatRules_Gt:
			v := float64(tt.Gt)
			schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
		case *validate.FloatRules_Gte:
			v := float64(tt.Gte)
			schema.Minimum = &v
		}
		if len(constraint.In) > 0 {
			items := make([]*yaml.Node, len(constraint.In))
			for i, item := range constraint.In {
				items[i] = utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 32))
			}
			schema.Enum = items
		}
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 32))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 32)))
	}
}

func updateSchemaDouble(schema *base.Schema, constraint *validate.DoubleRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatFloat(float64(*constraint.Const), 'f', -1, 64))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.DoubleRules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.DoubleRules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.DoubleRules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.DoubleRules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 64))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 64))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatFloat(float64(item), 'f', -1, 64)))
	}
}

func updateSchemaInt32(schema *base.Schema, constraint *validate.Int32Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(int64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Int32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Int32Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Int32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Int32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(int64(item), 10)))
	}
}

func updateSchemaInt64(schema *base.Schema, constraint *validate.Int64Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(int64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Int64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Int64Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Int64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Int64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(item, 10)))
	}
}

func updateSchemaUint32(schema *base.Schema, constraint *validate.UInt32Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatUint(uint64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.UInt32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.UInt32Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.UInt32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.UInt32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatUint(uint64(item), 10)))
	}
}

func updateSchemaUint64(schema *base.Schema, constraint *validate.UInt64Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatUint(uint64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.UInt64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.UInt64Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.UInt64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.UInt64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatUint(uint64(item), 10)))
	}
}

func updateSchemaSint32(schema *base.Schema, constraint *validate.SInt32Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(int64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SInt32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SInt32Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SInt32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SInt32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(int64(item), 10)))
	}
}

func updateSchemaSint64(schema *base.Schema, constraint *validate.SInt64Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(*constraint.Const, 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SInt64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SInt64Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SInt64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SInt64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(item, 10)))
	}
}

func updateSchemaFixed32(schema *base.Schema, constraint *validate.Fixed32Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatUint(uint64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Fixed32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Fixed32Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Fixed32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Fixed32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatUint(uint64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatUint(uint64(item), 10)))
	}
}

func updateSchemaFixed64(schema *base.Schema, constraint *validate.Fixed64Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(strconv.FormatUint(*constraint.Const, 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Fixed64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Fixed64Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Fixed64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.Fixed64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(strconv.FormatUint(item, 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(strconv.FormatUint(item, 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(strconv.FormatUint(item, 10)))
	}
}

func updateSchemaSfixed32(schema *base.Schema, constraint *validate.SFixed32Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(int64(*constraint.Const), 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SFixed32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SFixed32Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SFixed32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SFixed32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, OneOf: schema.OneOf, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(int64(item), 10)))
	}
}

func updateSchemaSfixed64(schema *base.Schema, constraint *validate.SFixed64Rules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(*constraint.Const, 10))
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SFixed64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SFixed64Rules_Lte:
		v := float64(tt.Lte)
		schema.Maximum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SFixed64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	case *validate.SFixed64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(item, 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(item, 10)))
	}
}

func updateSchemaBool(schema *base.Schema, constraint *validate.BoolRules) {
	if constraint.Const != nil {
		if *constraint.Const {
			schema.Const = utils.CreateStringNode("true")
		} else {
			schema.Const = utils.CreateStringNode("false")
		}
	}
	for _, item := range constraint.Example {
		value := "false"
		if item {
			value = "true"
		}
		schema.Examples = append(schema.Examples, utils.CreateBoolNode(value))
	}
}

//gocyclo:ignore
func updateSchemaString(schema *base.Schema, constraint *validate.StringRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(*constraint.Const)
	}
	if constraint.Len != nil {
		v := int64(*constraint.Len)
		schema.MaxLength = &v
		schema.MinLength = &v
	}
	if constraint.MinLen != nil {
		v := int64(*constraint.MinLen)
		schema.MinLength = &v
	}
	if constraint.MaxLen != nil {
		v := int64(*constraint.MaxLen)
		schema.MaxLength = &v
	}
	if constraint.Pattern != nil {
		schema.Pattern = *constraint.Pattern
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(item)
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(item)
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	switch v := constraint.WellKnown.(type) {
	case *validate.StringRules_Email:
		if v.Email {
			schema.Format = "email"
		}
	case *validate.StringRules_Hostname:
		if v.Hostname {
			schema.Format = "hostname"
		}
	case *validate.StringRules_Ip:
		if v.Ip {
			schema.Format = "ip"
		}
	case *validate.StringRules_Ipv4:
		if v.Ipv4 {
			schema.Format = "ipv4"
		}
	case *validate.StringRules_Ipv6:
		if v.Ipv6 {
			schema.Format = "ipv6"
		}
	case *validate.StringRules_Uri:
		if v.Uri {
			schema.Format = "uri"
		}
	case *validate.StringRules_UriRef:
		if v.UriRef {
			schema.Format = "uri-ref"
		}
	case *validate.StringRules_Address:
		if v.Address {
			schema.Format = "address"
		}
	case *validate.StringRules_Uuid:
		if v.Uuid {
			schema.Format = "uuid"
		}
	case *validate.StringRules_IpWithPrefixlen:
	case *validate.StringRules_Ipv4WithPrefixlen:
	case *validate.StringRules_Ipv6WithPrefixlen:
	case *validate.StringRules_IpPrefix:
	case *validate.StringRules_Ipv4Prefix:
	case *validate.StringRules_Ipv6Prefix:
	case *validate.StringRules_HostAndPort:
	case *validate.StringRules_WellKnownRegex:
		switch v.WellKnownRegex {
		case validate.KnownRegex_KNOWN_REGEX_HTTP_HEADER_NAME:
		case validate.KnownRegex_KNOWN_REGEX_HTTP_HEADER_VALUE:
		}
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(item))
	}
}

func updateSchemaBytes(schema *base.Schema, constraint *validate.BytesRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(string(constraint.Const))
	}
	if constraint.Len != nil {
		v := int64(*constraint.Len)
		schema.MaxLength = &v
		schema.MinLength = &v
	}
	if constraint.MinLen != nil {
		v := int64(*constraint.MinLen)
		schema.MinLength = &v
	}
	if constraint.MaxLen != nil {
		v := int64(*constraint.MaxLen)
		schema.MaxLength = &v
	}
	if constraint.Pattern != nil {
		schema.Pattern = *constraint.Pattern
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(string(item))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(string(item))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	switch v := constraint.WellKnown.(type) {
	case *validate.BytesRules_Ip:
		if v.Ip {
			schema.Format = "ip"
		}
	case *validate.BytesRules_Ipv4:
		if v.Ipv4 {
			schema.Format = "ipv4"
		}
	case *validate.BytesRules_Ipv6:
		if v.Ipv6 {
			schema.Format = "ipv6"
		}
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(string(item)))
	}
}

func updateSchemaEnum(schema *base.Schema, constraint *validate.EnumRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateIntNode(strconv.FormatInt(int64(*constraint.Const), 10))
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateIntNode(strconv.FormatInt(int64(item), 10))
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateIntNode(strconv.FormatInt(int64(item), 10)))
	}
}

func updateSchemaRepeated(schema *base.Schema, constraint *validate.RepeatedRules) {
	if constraint.Unique != nil {
		schema.UniqueItems = constraint.Unique
	}
	if constraint.MaxItems != nil {
		v := int64(*constraint.MaxItems)
		schema.MaxItems = &v
	}
	if constraint.MinItems != nil {
		v := int64(*constraint.MinItems)
		schema.MinItems = &v
	}
	if constraint.MaxItems != nil {
		v := int64(*constraint.MaxItems)
		schema.MaxItems = &v
	}
	if constraint.Items != nil && schema.Items != nil && schema.Items.A != nil && !schema.Items.A.IsReference() {
		updateSchemaWithFieldConstraints(schema.Items.A.Schema(), constraint.Items, false)
	}
}

func updateSchemaMap(schema *base.Schema, constraint *validate.MapRules) {
	if constraint.MinPairs != nil {
		v := int64(*constraint.MinPairs)
		schema.MinProperties = &v
	}
	if constraint.MaxPairs != nil {
		v := int64(*constraint.MaxPairs)
		schema.MaxProperties = &v
	}
	// NOTE: Most of these properties don't make sense for object keys
	// updateSchemaWithFieldConstraints(schema, constraint.Keys)
	if schema.AdditionalProperties != nil && constraint.Values != nil {
		updateSchemaWithFieldConstraints(schema.AdditionalProperties.A.Schema(), constraint.Values, false)
	}
}

func updateSchemaAny(schema *base.Schema, constraint *validate.AnyRules) {
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(item)
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(item)
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
}

func updateSchemaDuration(schema *base.Schema, constraint *validate.DurationRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(constraint.Const.String())
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(item.String())
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(item.String())
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(item.String()))
	}
}

func updateSchemaTimestamp(schema *base.Schema, constraint *validate.TimestampRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(constraint.Const.String())
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(item.String()))
	}
}
