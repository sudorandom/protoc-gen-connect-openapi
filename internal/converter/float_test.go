package converter_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestFloatSpecialValues(t *testing.T) {
	pf := loadTestFileDescriptorSet(t)
	opts, err := options.FromString("format=jsonschema")
	require.NoError(t, err)
	resp, err := converter.ConvertWithOptions(&pluginpb.CodeGeneratorRequest{
		ProtoFile:      pf.GetFile(),
		FileToGenerate: []string{"standard/float_special_values.proto"},
	}, opts)
	require.NoError(t, err)
	require.Len(t, resp.File, 1)
	compiler := compileJSONSchema(t, resp.File[0].GetContent())
	schema, err := compiler.Compile("schema.json#/$defs/test.special.v1.FloatValues")
	require.NoError(t, err)

	for _, number := range []float64{0, 1.25, math.NaN(), math.Inf(1), math.Inf(-1)} {
		encoded, err := protojson.Marshal(wrapperspb.Double(number))
		require.NoError(t, err)
		var value any
		require.NoError(t, json.Unmarshal(encoded, &value))
		for _, field := range []string{"floatValue", "doubleValue", "optionalFloat", "optionalDouble", "floatChoice", "doubleChoice"} {
			assert.NoError(t, schema.Validate(map[string]any{field: value}), "%s: %v", field, value)
		}
		assert.NoError(t, schema.Validate(map[string]any{
			"floats":    []any{value},
			"doubles":   []any{value},
			"floatMap":  map[string]any{"value": value},
			"doubleMap": map[string]any{"value": value},
		}), "container value: %v", value)
	}
	assert.NoError(t, schema.Validate(map[string]any{"optionalFloat": nil, "optionalDouble": nil}))
	for _, value := range []any{"nan", "inf", "1.25", "", true, nil} {
		assert.Error(t, schema.Validate(map[string]any{"doubleValue": value}), "invalid scalar: %v", value)
		assert.Error(t, schema.Validate(map[string]any{"floats": []any{value}}), "invalid list item: %v", value)
		assert.Error(t, schema.Validate(map[string]any{"doubleMap": map[string]any{"value": value}}), "invalid map value: %v", value)
	}

	constrained, err := compiler.Compile("schema.json#/$defs/test.special.v1.ConstrainedFloatValues")
	require.NoError(t, err)
	assert.NoError(t, constrained.Validate(map[string]any{
		"floatConst": 1.0, "doubleConst": 1.0, "doubleRange": 0.5,
		"floatFinite": 1.0, "doubleFinite": 1.0, "optionalFinite": nil,
		"annotatedRange": 0.5, "annotatedString": "custom", "annotatedNullable": nil,
		"floats": []any{1.0}, "doubles": map[string]any{"value": 1.0},
	}))
	for _, field := range []string{"floatConst", "doubleConst", "doubleRange", "annotatedRange"} {
		assert.Error(t, constrained.Validate(map[string]any{field: 2.0}), field)
	}
	for _, value := range []string{"NaN", "Infinity", "-Infinity"} {
		for _, field := range []string{"floatConst", "doubleConst", "doubleRange", "floatFinite", "doubleFinite", "optionalFinite", "annotatedRange", "annotatedNullable"} {
			assert.Error(t, constrained.Validate(map[string]any{field: value}), "%s: %s", field, value)
		}
		assert.Error(t, constrained.Validate(map[string]any{"floats": []any{value}}))
		assert.Error(t, constrained.Validate(map[string]any{"doubles": map[string]any{"value": value}}))
	}
}
