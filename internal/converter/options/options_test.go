package options_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
)

func TestFromString(t *testing.T) {
	t.Run("booleans", func(t *testing.T) {
		optionList := []string{
			"debug",
			"include-number-enum-values",
			"allow-get",
			"with-streaming",
			"with-proto-names",
			"with-proto-annotations",
			"trim-unused-types",
			"fully-qualified-message-names",
			"without-default-tags",
			"without-field-behavior-prefixes",
			"with-service-descriptions",
			"short-service-tags",
			"short-operation-ids",
			"with-google-error-detail",
		}
		opts, err := options.FromString(strings.Join(optionList, ","))
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
		assert.True(t, opts.WithoutFieldBehaviorPrefixes)
		assert.True(t, opts.WithServiceDescriptions)
		assert.True(t, opts.ShortServiceTags)
		assert.True(t, opts.ShortOperationIds)
		assert.True(t, opts.WithGoogleErrorDetail)

		t.Run("removed googleapi-http flags", func(t *testing.T) {
			for _, param := range []string{"only-googleapi-http", "ignore-googleapi-http"} {
				_, err := options.FromString(param)
				require.ErrorContains(t, err, "invalid parameter: "+param)
			}
		})
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
		assert.True(t, opts.WithoutFieldBehaviorPrefixes)
	})

	t.Run("features", func(t *testing.T) {
		t.Run("single feature", func(t *testing.T) {
			opts, err := options.FromString("features=connectrpc")
			require.NoError(t, err)
			assert.True(t, opts.FeatureEnabled(options.FeatureConnectRPC))
			assert.False(t, opts.FeatureEnabled(options.FeatureGoogleAPIHTTP))
		})

		t.Run("multiple features", func(t *testing.T) {
			opts, err := options.FromString("features=connectrpc;google.api.http")
			require.NoError(t, err)
			assert.True(t, opts.FeatureEnabled(options.FeatureConnectRPC))
			assert.True(t, opts.FeatureEnabled(options.FeatureGoogleAPIHTTP))
		})

		t.Run("invalid feature", func(t *testing.T) {
			_, err := options.FromString("features=invalid")
			require.Error(t, err)
		})

		t.Run("removed features", func(t *testing.T) {
			for _, feature := range []string{"twirp", "gnostic", "protovalidate"} {
				_, err := options.FromString("features=connectrpc;" + feature)
				require.ErrorContains(t, err, "invalid feature: '"+feature+"'")
			}
		})
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
		t.Run("jsonschema", func(t *testing.T) {
			opts, err := options.FromString("format=jsonschema")
			require.NoError(t, err)
			assert.Equal(t, options.FormatJSONSchema, opts.Format)
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
			assert.Equal(t, map[string]struct{}{"json": {}, "proto": {}}, opts.ContentTypes)
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

	t.Run("base and override options", func(t *testing.T) {
		t.Run("invalid base file extension", func(t *testing.T) {
			_, err := options.FromString("base=foo.txt")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "the file extension for 'base'")
		})

		t.Run("invalid override file extension", func(t *testing.T) {
			_, err := options.FromString("override=foo.txt")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "the file extension for 'override'")
		})

		t.Run("missing base file", func(t *testing.T) {
			_, err := options.FromString("base=nonexistent_file.yaml")
			require.Error(t, err)
		})

		t.Run("missing override file", func(t *testing.T) {
			_, err := options.FromString("override=nonexistent_file.yaml")
			require.Error(t, err)
		})
	})

	t.Run("asyncapi and allowed visibilities", func(t *testing.T) {
		opts, err := options.FromString("asyncapi-path=stream.yaml,asyncapi-channel-template=/ws/{service}/{method},allowed-visibilities=INTERNAL;PREVIEW")
		require.NoError(t, err)
		assert.Equal(t, "stream.yaml", opts.AsyncAPIPath)
		assert.Equal(t, "/ws/{service}/{method}", opts.AsyncAPIChannelTemplate)
		assert.True(t, opts.AllowedVisibilities["INTERNAL"])
		assert.True(t, opts.AllowedVisibilities["PREVIEW"])
		assert.False(t, opts.AllowedVisibilities["EXTERNAL"])
	})

	t.Run("well-known-type-descriptions", func(t *testing.T) {
		t.Run("defaults to concise", func(t *testing.T) {
			opts, err := options.FromString("")
			require.NoError(t, err)
			assert.Equal(t, options.WellKnownTypeDescriptionsConcise, opts.WellKnownTypeDescriptions)
		})

		for _, mode := range []options.WellKnownTypeDescriptionMode{
			options.WellKnownTypeDescriptionsFull,
			options.WellKnownTypeDescriptionsConcise,
			options.WellKnownTypeDescriptionsOmit,
		} {
			t.Run(string(mode), func(t *testing.T) {
				opts, err := options.FromString("well-known-type-descriptions=" + string(mode))
				require.NoError(t, err)
				assert.Equal(t, mode, opts.WellKnownTypeDescriptions)
			})
		}

		t.Run("invalid value", func(t *testing.T) {
			_, err := options.FromString("well-known-type-descriptions=brief")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "well-known-type-descriptions")
		})
	})

	t.Run("compat", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			opts, err := options.FromString("")
			require.NoError(t, err)
			assert.Equal(t, options.CompatNone, opts.Compat)
		})

		t.Run("openapi-3.0", func(t *testing.T) {
			opts, err := options.FromString("compat=openapi-3.0")
			require.NoError(t, err)
			assert.Equal(t, options.CompatOpenAPI30, opts.Compat)
		})

		t.Run("invalid value", func(t *testing.T) {
			_, err := options.FromString("compat=swagger-2.0")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "compat")
		})
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

func TestIsValidContentType(t *testing.T) {
	assert.True(t, options.IsValidContentType("json"))
	assert.True(t, options.IsValidContentType("proto"))
	assert.True(t, options.IsValidContentType("connect+json"))
	assert.True(t, options.IsValidContentType("connect+proto"))
	assert.False(t, options.IsValidContentType("invalid_type"))
}

func TestCompileServicePatterns(t *testing.T) {
	t.Run("valid patterns", func(t *testing.T) {
		patterns, err := options.CompileServicePatterns([]string{"foo.*", "bar.**"})
		require.NoError(t, err)
		assert.Len(t, patterns, 2)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := options.CompileServicePatterns([]string{"["})
		require.Error(t, err)
	})
}

func TestGetExtensionTypeResolver(t *testing.T) {
	opts := options.NewOptions()
	assert.NotNil(t, opts.GetExtensionTypeResolver())
}
