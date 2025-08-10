package protovalidate

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"buf.build/go/protovalidate"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/utils"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

func SchemaWithMessageAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	rules, err := protovalidate.ResolveMessageRules(desc)
	if err != nil {
		slog.Warn("unable to resolve message rules", slog.Any("error", err))
		return schema
	}
	if rules == nil {
		return schema
	}
	updateWithCEL(schema, rules.GetCel(), nil, nil)
	return schema
}

func SchemaWithFieldAnnotations(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	rules, err := protovalidate.ResolveFieldRules(desc)
	if err != nil {
		slog.Warn("unable to resolve field rules", slog.Any("error", err))
		return schema
	}
	if rules == nil {
		return schema
	}

	updateWithCEL(schema, rules.GetCel(), nil, nil)
	if rules.Required != nil && *rules.Required {
		parent := schema.ParentProxy.Schema()
		if parent != nil {
			parent.Required = util.AppendStringDedupe(parent.Required, util.MakeFieldName(opts, desc))
		}
	}
	updateSchemaWithFieldRules(opts, schema, rules, onlyScalar)
	return schema
}

func PopulateParentProperties(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	if parent == nil {
		return parent
	}
	rules, err := protovalidate.ResolveFieldRules(desc)
	if err != nil {
		slog.Warn("unable to resolve field rules", slog.Any("error", err))
		return parent
	}
	if rules == nil {
		return parent
	}
	if rules.Required != nil && *rules.Required {
		parent.Required = util.AppendStringDedupe(parent.Required, util.MakeFieldName(opts, desc))
	}
	return parent
}

//gocyclo:ignore
func updateSchemaWithFieldRules(opts options.Options, schema *base.Schema, rules *validate.FieldRules, onlyScalar bool) {
	if rules == nil {
		return
	}

	var innerRules protoreflect.Message
	switch t := rules.Type.(type) {
	case *validate.FieldRules_Float:
		innerRules = t.Float.ProtoReflect()
		updateSchemaFloat(schema, t.Float)
	case *validate.FieldRules_Double:
		innerRules = t.Double.ProtoReflect()
		updateSchemaDouble(schema, t.Double)
	case *validate.FieldRules_Int32:
		innerRules = t.Int32.ProtoReflect()
		updateSchemaInt32(schema, t.Int32)
	case *validate.FieldRules_Int64:
		innerRules = t.Int64.ProtoReflect()
		updateSchemaInt64(schema, t.Int64)
	case *validate.FieldRules_Uint32:
		innerRules = t.Uint32.ProtoReflect()
		updateSchemaUint32(schema, t.Uint32)
	case *validate.FieldRules_Uint64:
		innerRules = t.Uint64.ProtoReflect()
		updateSchemaUint64(schema, t.Uint64)
	case *validate.FieldRules_Sint32:
		innerRules = t.Sint32.ProtoReflect()
		updateSchemaSint32(schema, t.Sint32)
	case *validate.FieldRules_Sint64:
		innerRules = t.Sint64.ProtoReflect()
		updateSchemaSint64(schema, t.Sint64)
	case *validate.FieldRules_Fixed32:
		innerRules = t.Fixed32.ProtoReflect()
		updateSchemaFixed32(schema, t.Fixed32)
	case *validate.FieldRules_Fixed64:
		innerRules = t.Fixed64.ProtoReflect()
		updateSchemaFixed64(schema, t.Fixed64)
	case *validate.FieldRules_Sfixed32:
		innerRules = t.Sfixed32.ProtoReflect()
		updateSchemaSfixed32(schema, t.Sfixed32)
	case *validate.FieldRules_Sfixed64:
		innerRules = t.Sfixed64.ProtoReflect()
		updateSchemaSfixed64(schema, t.Sfixed64)
	case *validate.FieldRules_Bool:
		innerRules = t.Bool.ProtoReflect()
		updateSchemaBool(schema, t.Bool)
	case *validate.FieldRules_String_:
		innerRules = t.String_.ProtoReflect()
		updateSchemaString(schema, t.String_)
	case *validate.FieldRules_Bytes:
		innerRules = t.Bytes.ProtoReflect()
		updateSchemaBytes(schema, t.Bytes)
	case *validate.FieldRules_Enum:
		innerRules = t.Enum.ProtoReflect()
		updateSchemaEnum(schema, t.Enum)
	case *validate.FieldRules_Any:
		innerRules = t.Any.ProtoReflect()
		updateSchemaAny(schema, t.Any)
	case *validate.FieldRules_Duration:
		innerRules = t.Duration.ProtoReflect()
		updateSchemaDuration(schema, t.Duration)
	case *validate.FieldRules_Timestamp:
		innerRules = t.Timestamp.ProtoReflect()
		updateSchemaTimestamp(schema, t.Timestamp)
	}

	if innerRules != nil {
		if err := reparseUnrecognized(opts.GetExtensionTypeResolver(), innerRules); err != nil {
			slog.Error("failed to reparse unrecognized fields", "error", err)
		}

		// Collect rules to sort them for consistent output
		type ruleEntry struct {
			fd protoreflect.FieldDescriptor
			v  protoreflect.Value
		}
		var ruleEntries []ruleEntry

		innerRules.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			ruleEntries = append(ruleEntries, ruleEntry{fd: fd, v: v})
			return true
		})

		// Sort rules by field name for consistent output
		slices.SortFunc(ruleEntries, func(a, b ruleEntry) int {
			return strings.Compare(string(a.fd.Name()), string(b.fd.Name()))
		})

		for _, entry := range ruleEntries {
			predefinedRules, err := protovalidate.ResolvePredefinedRules(entry.fd)
			if err != nil {
				slog.Error("error resolving predefined rules", "error", err)
				continue
			}
			if predefinedRules == nil {
				continue
			}
			updateWithCEL(schema, predefinedRules.GetCel(), &entry.v, entry.fd)
		}
	}

	if !onlyScalar {
		switch t := rules.Type.(type) {
		case *validate.FieldRules_Repeated:
			updateSchemaRepeated(opts, schema, t.Repeated)
		case *validate.FieldRules_Map:
			updateSchemaMap(opts, schema, t.Map)
		}
	}
}

func updateWithCEL(schema *base.Schema, rules []*validate.Rule, val *protoreflect.Value, fieldDesc protoreflect.FieldDescriptor) {
	if len(rules) == 0 {
		return
	}
	b := strings.Builder{}
	if schema.Description != "" {
		b.WriteString(strings.TrimSpace(schema.Description))
		b.WriteByte('\n')
	}

	slices.SortFunc(rules, func(a, b *validate.Rule) int {
		return strings.Compare(a.GetId(), b.GetId())
	})

	for _, cel := range rules {
		if cel.HasId() {
			b.WriteString(cel.GetId())
		}
		if val != nil {
			b.WriteString(" = ")
			b.WriteString(formatProtoreflectValue(*val, fieldDesc))
		}
		if cel.HasMessage() {
			b.WriteString(" // ")
			b.WriteString(cel.GetMessage())
		}
		// This is excluded because it's very verbose.
		// if cel.HasExpression() {
		// 	b.WriteString("\n```\n")
		// 	b.WriteString(cel.GetExpression())
		// 	b.WriteString("\n```\n")
		// }
		b.WriteString("\n")
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

func updateSchemaRepeated(opts options.Options, schema *base.Schema, constraint *validate.RepeatedRules) {
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
		updateSchemaWithFieldRules(opts, schema.Items.A.Schema(), constraint.Items, false)
	}
}

func updateSchemaMap(opts options.Options, schema *base.Schema, constraint *validate.MapRules) {
	if constraint.MinPairs != nil {
		v := int64(*constraint.MinPairs)
		schema.MinProperties = &v
	}
	if constraint.MaxPairs != nil {
		v := int64(*constraint.MaxPairs)
		schema.MaxProperties = &v
	}
	// NOTE: Most of these properties don't make sense for object keys
	// updateSchemaWithFieldRules(schema, constraint.Keys)
	if schema.AdditionalProperties != nil && constraint.Values != nil {
		updateSchemaWithFieldRules(opts, schema.AdditionalProperties.A.Schema(), constraint.Values, false)
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
		schema.Const = utils.CreateStringNode(constraint.Const.AsDuration().String())
	}
	if len(constraint.In) > 0 {
		items := make([]*yaml.Node, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = utils.CreateStringNode(item.AsDuration().String())
		}
		schema.Enum = items
	}
	if len(constraint.NotIn) > 0 {
		items := make([]*yaml.Node, len(constraint.NotIn))
		for i, item := range constraint.NotIn {
			items[i] = utils.CreateStringNode(item.AsDuration().String())
		}
		schema.Not = base.CreateSchemaProxy(&base.Schema{Type: schema.Type, Enum: items})
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(item.AsDuration().String()))
	}
}

func updateSchemaTimestamp(schema *base.Schema, constraint *validate.TimestampRules) {
	if constraint.Const != nil {
		schema.Const = utils.CreateStringNode(constraint.Const.AsTime().String())
	}
	for _, item := range constraint.Example {
		schema.Examples = append(schema.Examples, utils.CreateStringNode(item.AsTime().String()))
	}
}

func reparseUnrecognized(
	extensionTypeResolver protoregistry.ExtensionTypeResolver,
	reflectMessage protoreflect.Message,
) error {
	if unknown := reflectMessage.GetUnknown(); len(unknown) > 0 {
		reflectMessage.SetUnknown(nil)
		options := proto.UnmarshalOptions{
			Resolver: extensionTypeResolver,
			Merge:    true,
		}
		if err := options.Unmarshal(unknown, reflectMessage.Interface()); err != nil {
			return err
		}
	}
	return nil
}

func formatProtoreflectValue(val protoreflect.Value, fieldDesc protoreflect.FieldDescriptor) string {
	switch v := val.Interface().(type) {
	case protoreflect.List:
		var elements []string
		for i := 0; i < v.Len(); i++ {
			elements = append(elements, formatProtoreflectValue(v.Get(i), fieldDesc))
		}
		return "[" + strings.Join(elements, ", ") + "]"
	case protoreflect.EnumNumber:
		if fieldDesc != nil && fieldDesc.Kind() == protoreflect.EnumKind {
			enumDesc := fieldDesc.Enum()
			enumValDesc := enumDesc.Values().ByNumber(v)
			if enumValDesc != nil {
				return string(enumValDesc.Name())
			}
		}
		return strconv.Itoa(int(v))
	case string:
		return strconv.Quote(v)
	case *durationpb.Duration:
		return v.AsDuration().String()
	case *timestamppb.Timestamp:
		// RFC3339Nano for OpenAPI compatibility
		return v.AsTime().Format(time.RFC3339Nano)
	case proto.Message:
		// Special handling for Duration and Timestamp proto messages
		switch msg := v.(type) {
		case *durationpb.Duration:
			return msg.AsDuration().String()
		case *timestamppb.Timestamp:
			return msg.AsTime().Format(time.RFC3339Nano)
		default:
			data, _ := protojson.Marshal(msg)
			return string(data)
		}
	case protoreflect.Message:
		// Try to convert to well-known types
		if m, ok := v.Interface().(*durationpb.Duration); ok {
			return m.AsDuration().String()
		}
		if m, ok := v.Interface().(*timestamppb.Timestamp); ok {
			return m.AsTime().Format(time.RFC3339Nano)
		}
		data, _ := protojson.Marshal(val.Message().Interface())
		return string(data)
	}
	return val.String()
}
