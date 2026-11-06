package gnostic

import (
	"testing"

	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToLinks(t *testing.T) {
	t.Parallel()
	links := toLinks(&goa3.LinksOrReferences{
		AdditionalProperties: []*goa3.NamedLinkOrReference{
			{
				Name: "byRef",
				Value: &goa3.LinkOrReference{Oneof: &goa3.LinkOrReference_Reference{
					Reference: &goa3.Reference{XRef: "#/components/links/Other"},
				}},
			},
			{Name: "empty", Value: &goa3.LinkOrReference{}},
		},
	})
	require.Equal(t, 1, links.Len())
	link, ok := links.Get("byRef")
	require.True(t, ok)
	assert.Equal(t, "#/components/links/Other", link.Reference)
}

func TestToDefault(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		input    *goa3.DefaultType
		expected string
	}{
		{"integer", &goa3.DefaultType{Oneof: &goa3.DefaultType_Number{Number: 5}}, "5\n"},
		{"float", &goa3.DefaultType{Oneof: &goa3.DefaultType_Number{Number: 1.5}}, "1.5\n"},
		{"boolean", &goa3.DefaultType{Oneof: &goa3.DefaultType_Boolean{Boolean: true}}, "true\n"},
		{"string", &goa3.DefaultType{Oneof: &goa3.DefaultType_String_{String_: "5"}}, "\"5\"\n"},
	} {
		out, err := yaml.Marshal(toDefault(tc.input))
		require.NoError(t, err)
		assert.Equal(t, tc.expected, string(out), tc.name)
	}
}
