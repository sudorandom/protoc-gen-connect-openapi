package protovalidate

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestUpdateSchemaRules(t *testing.T) {
	opts := options.NewOptions()

	t.Run("duration gte and lte", func(t *testing.T) {
		schema := &base.Schema{}
		rules := &validate.FieldRules{
			Type: &validate.FieldRules_Duration{
				Duration: &validate.DurationRules{
					GreaterThan: &validate.DurationRules_Gte{
						Gte: &durationpb.Duration{Seconds: 60},
					},
					LessThan: &validate.DurationRules_Lte{
						Lte: &durationpb.Duration{Seconds: 3600},
					},
				},
			},
		}
		rulesClone := proto.Clone(rules).(*validate.FieldRules)
		updateSchemaWithFieldRules(opts, schema, rulesClone, false, nil)

		assert.Contains(t, schema.Description, "duration.gte = 1m0s")
		assert.Contains(t, schema.Description, "duration.lte = 1h0m0s")
		assert.NotContains(t, schema.Description, "gte_lt")
		assert.NotContains(t, schema.Description, "gte_lte")
	})

	t.Run("duration gt and lt", func(t *testing.T) {
		schema := &base.Schema{}
		rules := &validate.FieldRules{
			Type: &validate.FieldRules_Duration{
				Duration: &validate.DurationRules{
					GreaterThan: &validate.DurationRules_Gt{
						Gt: &durationpb.Duration{Seconds: 10},
					},
					LessThan: &validate.DurationRules_Lt{
						Lt: &durationpb.Duration{Seconds: 20},
					},
				},
			},
		}
		rulesClone := proto.Clone(rules).(*validate.FieldRules)
		updateSchemaWithFieldRules(opts, schema, rulesClone, false, nil)

		assert.Contains(t, schema.Description, "duration.gt = 10s")
		assert.Contains(t, schema.Description, "duration.lt = 20s")
		assert.NotContains(t, schema.Description, "gt_lt")
	})

	t.Run("field_mask in and not_in", func(t *testing.T) {
		schema := &base.Schema{}
		rules := &validate.FieldRules{
			Type: &validate.FieldRules_FieldMask{
				FieldMask: &validate.FieldMaskRules{
					In:    []string{"name"},
					NotIn: []string{"secret"},
				},
			},
		}
		rulesClone := proto.Clone(rules).(*validate.FieldRules)
		updateSchemaWithFieldRules(opts, schema, rulesClone, false, nil)

		assert.Contains(t, schema.Description, "field_mask.in")
		assert.Contains(t, schema.Description, `["name"]`)
		assert.Contains(t, schema.Description, "field_mask.not_in")
		assert.Contains(t, schema.Description, `["secret"]`)
	})

	t.Run("enum defined_only", func(t *testing.T) {
		schema := &base.Schema{}
		rules := &validate.FieldRules{
			Type: &validate.FieldRules_Enum{
				Enum: &validate.EnumRules{
					DefinedOnly: proto.Bool(true),
				},
			},
		}
		rulesClone := proto.Clone(rules).(*validate.FieldRules)
		updateSchemaWithFieldRules(opts, schema, rulesClone, false, nil)

		require.NotEmpty(t, schema.Description)
		assert.Contains(t, schema.Description, "enum.defined_only = true")
	})

	t.Run("enum defined_only false", func(t *testing.T) {
		schema := &base.Schema{}
		rules := &validate.FieldRules{
			Type: &validate.FieldRules_Enum{
				Enum: &validate.EnumRules{
					DefinedOnly: proto.Bool(false),
				},
			},
		}
		rulesClone := proto.Clone(rules).(*validate.FieldRules)
		updateSchemaWithFieldRules(opts, schema, rulesClone, false, nil)

		assert.Empty(t, schema.Description)
	})
}
