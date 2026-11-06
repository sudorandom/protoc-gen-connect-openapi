package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperSnakeCase(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"Color":       "COLOR",
		"ColorMode":   "COLOR_MODE",
		"HTTPMethod":  "HTTP_METHOD",
		"V2Type":      "V2_TYPE",
		"Already_Ok":  "ALREADY_OK",
		"ID":          "ID",
		"SimpleHTTP":  "SIMPLE_HTTP",
		"GPUKindType": "GPU_KIND_TYPE",
	}
	for input, expected := range tests {
		assert.Equal(t, expected, upperSnakeCase(input), input)
	}
}
