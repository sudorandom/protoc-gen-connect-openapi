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

package protovalidate

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/stretchr/testify/assert"
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

		assert.Contains(t, schema.Description, "duration.gte = 60s")
		assert.Contains(t, schema.Description, "duration.lte = 3600s")
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

		// The enum schema already restricts values to the defined ones.
		assert.Empty(t, schema.Description)
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
