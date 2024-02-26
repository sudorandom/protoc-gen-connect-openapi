package protovalidate

import (
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/bufbuild/protovalidate-go/resolver"
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SchemaWithMessageAnnotations(schema *jsonschema.Schema, desc protoreflect.MessageDescriptor) *jsonschema.Schema {
	r := resolver.DefaultResolver{}
	constraints := r.ResolveMessageConstraints(desc)
	if constraints == nil || constraints.GetDisabled() {
		return schema
	}
	updateWithCEL(schema, constraints.GetCel())
	return schema
}

func SchemaWithFieldAnnotations(schema *jsonschema.Schema, desc protoreflect.FieldDescriptor) *jsonschema.Schema {
	r := resolver.DefaultResolver{}
	constraints := r.ResolveFieldConstraints(desc)
	if constraints == nil {
		return schema
	}
	updateWithCEL(schema, constraints.GetCel())
	if constraints.Required {
		schema.Parent.Required = append(schema.Parent.Required, desc.JSONName())
	}
	updateSchemaWithFieldConstraints(schema, constraints)
	return schema
}

func updateSchemaWithFieldConstraints(schema *jsonschema.Schema, constraints *validate.FieldConstraints) {
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
	case *validate.FieldConstraints_Repeated:
		updateSchemaRepeated(schema, t.Repeated)
	case *validate.FieldConstraints_Map:
		updateSchemaMap(schema, t.Map)
	case *validate.FieldConstraints_Any:
		updateSchemaAny(schema, t.Any)
	case *validate.FieldConstraints_Duration:
		updateSchemaDuration(schema, t.Duration)
	case *validate.FieldConstraints_Timestamp:
		updateSchemaTimestamp(schema, t.Timestamp)
	}
}

func updateWithCEL(schema *jsonschema.Schema, constraints []*validate.Constraint) {
	if len(constraints) == 0 {
		return
	}
	b := strings.Builder{}
	if schema.Description != nil && *schema.Description != "" {
		b.WriteString(*schema.Description)
		b.WriteByte('\n')
	}
	for _, cel := range constraints {
		b.WriteString(cel.Message)
		b.WriteString(":\n```\n")
		b.WriteString(cel.Expression)
		b.WriteString("\n```\n\n")
	}
	s := b.String()
	schema.Description = &s
}

func updateSchemaFloat(schema *jsonschema.Schema, constraint *validate.FloatRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.FloatRules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.FloatRules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.FloatRules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.FloatRules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaDouble(schema *jsonschema.Schema, constraint *validate.DoubleRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.DoubleRules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.DoubleRules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.DoubleRules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.DoubleRules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaInt32(schema *jsonschema.Schema, constraint *validate.Int32Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Int32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.Int32Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Int32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.Int32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaInt64(schema *jsonschema.Schema, constraint *validate.Int64Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Int64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.Int64Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Int64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.Int64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaUint32(schema *jsonschema.Schema, constraint *validate.UInt32Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.UInt32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.UInt32Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.UInt32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.UInt32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaUint64(schema *jsonschema.Schema, constraint *validate.UInt64Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.UInt64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.UInt64Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.UInt64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.UInt64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaSint32(schema *jsonschema.Schema, constraint *validate.SInt32Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SInt32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.SInt32Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SInt32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.SInt32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaSint64(schema *jsonschema.Schema, constraint *validate.SInt64Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SInt64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.SInt64Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SInt64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.SInt64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaFixed32(schema *jsonschema.Schema, constraint *validate.Fixed32Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Fixed32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.Fixed32Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Fixed32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.Fixed32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaFixed64(schema *jsonschema.Schema, constraint *validate.Fixed64Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.Fixed64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.Fixed64Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.Fixed64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.Fixed64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaSfixed32(schema *jsonschema.Schema, constraint *validate.SFixed32Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SFixed32Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.SFixed32Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SFixed32Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.SFixed32Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaSfixed64(schema *jsonschema.Schema, constraint *validate.SFixed64Rules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	switch tt := constraint.LessThan.(type) {
	case *validate.SFixed64Rules_Lt:
		v := float64(tt.Lt)
		schema.ExclusiveMinimum = &v
	case *validate.SFixed64Rules_Lte:
		v := float64(tt.Lte)
		schema.Minimum = &v
	}
	switch tt := constraint.GreaterThan.(type) {
	case *validate.SFixed64Rules_Gt:
		v := float64(tt.Gt)
		schema.ExclusiveMinimum = &v
	case *validate.SFixed64Rules_Gte:
		v := float64(tt.Gte)
		schema.Minimum = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaBool(schema *jsonschema.Schema, constraint *validate.BoolRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
}

func updateSchemaString(schema *jsonschema.Schema, constraint *validate.StringRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	if constraint.Len != nil {
		v := int64(*constraint.Len)
		schema.MaxLength = &v
		schema.MinLength = v
	}
	if constraint.MinLen != nil {
		schema.MinLength = int64(*constraint.MinLen)
	}
	if constraint.MaxLen != nil {
		v := int64(*constraint.MaxLen)
		schema.MaxLength = &v
	}
	if constraint.Pattern != nil {
		schema.Pattern = constraint.Pattern
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
	switch v := constraint.WellKnown.(type) {
	case *validate.StringRules_Email:
		if v.Email {
			v := "email"
			schema.Format = &v
		}
	case *validate.StringRules_Hostname:
		if v.Hostname {
			v := "hostname"
			schema.Format = &v
		}
	case *validate.StringRules_Ip:
		if v.Ip {
			v := "ip"
			schema.Format = &v
		}
	case *validate.StringRules_Ipv4:
		if v.Ipv4 {
			v := "ipv4"
			schema.Format = &v
		}
	case *validate.StringRules_Ipv6:
		if v.Ipv6 {
			v := "ipv6"
			schema.Format = &v
		}
	case *validate.StringRules_Uri:
		if v.Uri {
			v := "uri"
			schema.Format = &v
		}
	case *validate.StringRules_UriRef:
		if v.UriRef {
			v := "uri-ref"
			schema.Format = &v
		}
	case *validate.StringRules_Address:
		if v.Address {
			v := "address"
			schema.Format = &v
		}
	case *validate.StringRules_Uuid:
		if v.Uuid {
			v := "uuid"
			schema.Format = &v
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
}

func updateSchemaBytes(schema *jsonschema.Schema, constraint *validate.BytesRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	if constraint.Len != nil {
		v := int64(*constraint.Len)
		schema.MaxLength = &v
		schema.MinLength = v
	}
	if constraint.MinLen != nil {
		schema.MinLength = int64(*constraint.MinLen)
	}
	if constraint.MaxLen != nil {
		v := int64(*constraint.MaxLen)
		schema.MaxLength = &v
	}
	if constraint.Pattern != nil {
		schema.Pattern = constraint.Pattern
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
	switch v := constraint.WellKnown.(type) {
	case *validate.BytesRules_Ip:
		if v.Ip {
			v := "ip"
			schema.Format = &v
		}
	case *validate.BytesRules_Ipv4:
		if v.Ipv4 {
			v := "ipv4"
			schema.Format = &v
		}
	case *validate.BytesRules_Ipv6:
		if v.Ipv6 {
			v := "ipv6"
			schema.Format = &v
		}
	}
}

func updateSchemaEnum(schema *jsonschema.Schema, constraint *validate.EnumRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaRepeated(schema *jsonschema.Schema, constraint *validate.RepeatedRules) {
	if constraint.Unique != nil {
		schema.UniqueItems = constraint.Unique
	}
	if constraint.MaxItems != nil {
		v := int64(*constraint.MaxItems)
		schema.MaxItems = &v
	}
	if constraint.MinItems != nil {
		schema.MinItems = int64(*constraint.MinItems)
	}
	if constraint.MaxItems != nil {
		v := int64(*constraint.MaxItems)
		schema.MaxItems = &v
	}
	if constraint.Items != nil && schema.Items != nil {
		for _, item := range schema.Items.SchemaArray {
			if item.TypeObject == nil {
				continue
			}
			updateSchemaWithFieldConstraints(item.TypeObject, constraint.Items)
		}
	}
}

func updateSchemaMap(schema *jsonschema.Schema, constraint *validate.MapRules) {
	if constraint.MinPairs != nil {
		schema.MinItems = int64(*constraint.MinPairs)
	}
	if constraint.MaxPairs != nil {
		v := int64(*constraint.MaxPairs)
		schema.MaxItems = &v
	}
	updateSchemaWithFieldConstraints(schema, constraint.Keys)
	if schema.AdditionalItems != nil && constraint.Values != nil {
		updateSchemaWithFieldConstraints(schema.AdditionalItems.TypeObject, constraint.Values)
	}
}

func updateSchemaAny(schema *jsonschema.Schema, constraint *validate.AnyRules) {
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaDuration(schema *jsonschema.Schema, constraint *validate.DurationRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
	if len(constraint.In) > 0 {
		items := make([]interface{}, len(constraint.In))
		for i, item := range constraint.In {
			items[i] = item
		}
		schema.Enum = items
	}
}

func updateSchemaTimestamp(schema *jsonschema.Schema, constraint *validate.TimestampRules) {
	if constraint.Const != nil {
		v := interface{}(constraint.Const)
		schema.Const = &v
	}
}
