package options_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
)

func TestFromString(t *testing.T) {
	t.Run("booleans", func(t *testing.T) {
		opts, err := options.FromString("debug,include-number-enum-values,allow-get,with-streaming,with-proto-names,with-proto-annotations,trim-unused-types,fully-qualified-message-names,without-default-tags,with-service-descriptions,ignore-googleapi-http,only-googleapi-http,short-service-tags,short-operation-ids,with-google-error-detail")
		require.NoError(t, err)
		assert.True(t, opts.Debug)
		assert.True(t, opts.IncludeNumberEnumValues)
		assert.True(t, opts.AllowGET)
		assert.True(t, opts.WithStreaming)
		assert.True(t, opts.WithProtoNames)
		assert.True(t, opts.WithProtoAnnotations)
		assert.True(t, opts.TrimUnusedTypes)
		assert.True(t, opts.FullyQualifiedMessageNames)
		assert.True(t, opts.WithoutDefaultTags)
		assert.True(t, opts.WithServiceDescriptions)
		assert.True(t, opts.IgnoreGoogleapiHTTP)
		assert.True(t, opts.OnlyGoogleapiHTTP)
		assert.True(t, opts.ShortServiceTags)
		assert.True(t, opts.ShortOperationIds)
		assert.True(t, opts.WithGoogleErrorDetail)
	})

	t.Run("format", func(t *testing.T) {
		t.Run("yaml", func(t *testing.T) {
			opts, err := options.FromString("format=yaml")
			require.NoError(t, err)
			assert.Equal(t, "yaml", opts.Format)
		})
		t.Run("json", func(t *testing.T) {
			opts, err := options.FromString("format=json")
			require.NoError(t, err)
			assert.Equal(t, "json", opts.Format)
		})
		t.Run("invalid", func(t *testing.T) {
			_, err := options.FromString("format=invalid")
			require.Error(t, err)
		})
	})

	t.Run("path", func(t *testing.T) {
		opts, err := options.FromString("path=/tmp/openapi.yaml")
		require.NoError(t, err)
		assert.Equal(t, "/tmp/openapi.yaml", opts.Path)
	})

	t.Run("path-prefix", func(t *testing.T) {
		opts, err := options.FromString("path-prefix=/api/v1")
		require.NoError(t, err)
		assert.Equal(t, "/api/v1", opts.PathPrefix)
	})

	t.Run("services", func(t *testing.T) {
		opts, err := options.FromString("services=foo.v1.FooService,services=bar.v1.BarService")
		require.NoError(t, err)
		assert.Len(t, opts.Services, 2)
	})

	t.Run("content-types", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			opts, err := options.FromString("")
			require.NoError(t, err)
			assert.Equal(t, map[string]struct{}{"json": {}}, opts.ContentTypes)
		})
		t.Run("custom", func(t *testing.T) {
			opts, err := options.FromString("content-types=proto")
			require.NoError(t, err)
			assert.Equal(t, map[string]struct{}{"proto": {}}, opts.ContentTypes)
		})
		t.Run("multiple", func(t *testing.T) {
			opts, err := options.FromString("content-types=json;proto")
			require.NoError(t, err)
			assert.Equal(t, map[string]struct{}{ "json": {}, "proto": {}}, opts.ContentTypes)
		})
		t.Run("invalid", func(t *testing.T) {
			_, err := options.FromString("content-types=invalid")
			require.Error(t, err)
		})
	})

	t.Run("invalid param", func(t *testing.T) {
		_, err := options.FromString("invalid-param")
		require.Error(t, err)
	})
}

func TestHasService(t *testing.T) {
	t.Run("no services configured", func(t *testing.T) {
		opts := options.NewOptions()
		assert.True(t, opts.HasService("any.service"))
	})

	t.Run("with services configured", func(t *testing.T) {
		opts, err := options.FromString("services=foo.v1.FooService,services=bar.v1.*")
		require.NoError(t, err)

		assert.True(t, opts.HasService("foo.v1.FooService"))
		assert.True(t, opts.HasService("bar.v1.BarService"))
		assert.True(t, opts.HasService("bar.v1.AnotherService"))
		assert.False(t, opts.HasService("baz.v1.BazService"))
	})
}