package util

import (
	"testing"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleRPCErrorDetailSchemasFieldNames(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		opts     options.Options
		expected string
	}{
		"json names":  {opts: options.Options{}, expected: "fieldViolations"},
		"proto names": {opts: options.Options{WithProtoNames: true}, expected: "field_violations"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			badRequest, ok := GoogleRPCErrorDetailSchemas(test.opts).Get("google.rpc.BadRequest")
			require.True(t, ok)
			_, ok = badRequest.Schema().Properties.Get(test.expected)
			assert.True(t, ok, "missing property %q", test.expected)
		})
	}
}
