package visibility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api_visibility "google.golang.org/genproto/googleapis/api/visibility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestShouldBeFiltered(t *testing.T) {
	tests := []struct {
		name       string
		rule       *api_visibility.VisibilityRule
		selectors  map[string]bool
		wantFilter bool
	}{
		{
			name:       "nil rule is never filtered",
			rule:       nil,
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "rule with empty selectors is always filtered",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL"},
			selectors:  map[string]bool{},
			wantFilter: true,
		},
		{
			name:       "single restriction matching selector",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "single restriction not matching selector",
			rule:       &api_visibility.VisibilityRule{Restriction: "EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: true,
		},
		{
			name:       "comma-separated restriction with first label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with second label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"EXTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with no label matching",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL,EXTERNAL"},
			selectors:  map[string]bool{"PREVIEW": true},
			wantFilter: true,
		},
		{
			name:       "comma-separated restriction with spaces",
			rule:       &api_visibility.VisibilityRule{Restriction: "INTERNAL, EXTERNAL"},
			selectors:  map[string]bool{"EXTERNAL": true},
			wantFilter: false,
		},
		{
			name:       "comma-separated restriction with multiple selectors",
			rule:       &api_visibility.VisibilityRule{Restriction: "PREVIEW,EXTERNAL"},
			selectors:  map[string]bool{"INTERNAL": true, "PREVIEW": true},
			wantFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldBeFiltered(tt.rule, tt.selectors)
			if got != tt.wantFilter {
				t.Errorf("ShouldBeFiltered() = %v, want %v", got, tt.wantFilter)
			}
		})
	}
}

type mockDesc struct {
	protoreflect.Descriptor
}

type mockFieldDesc struct {
	protoreflect.FieldDescriptor
	opts *descriptorpb.FieldOptions
}

func (m *mockFieldDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

type mockMessageDesc struct {
	protoreflect.MessageDescriptor
	opts *descriptorpb.MessageOptions
}

func (m *mockMessageDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

type mockMethodDesc struct {
	protoreflect.MethodDescriptor
	opts *descriptorpb.MethodOptions
}

func (m *mockMethodDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

type mockServiceDesc struct {
	protoreflect.ServiceDescriptor
	opts *descriptorpb.ServiceOptions
}

func (m *mockServiceDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

type mockEnumDesc struct {
	protoreflect.EnumDescriptor
	opts *descriptorpb.EnumOptions
}

func (m *mockEnumDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

type mockEnumValueDesc struct {
	protoreflect.EnumValueDescriptor
	opts *descriptorpb.EnumValueOptions
}

func (m *mockEnumValueDesc) Options() protoreflect.ProtoMessage {
	return m.opts
}

func TestGetVisibilityRuleAndExtract(t *testing.T) {
	t.Run("nil descriptor or unsupported descriptor", func(t *testing.T) {
		assert.Nil(t, GetVisibilityRule(mockDesc{}))
		assert.Equal(t, "", ExtractVisibilityRestriction(mockDesc{}))
	})

	t.Run("field with and without visibility", func(t *testing.T) {
		fdWithout := &mockFieldDesc{opts: &descriptorpb.FieldOptions{}}
		assert.Nil(t, GetVisibilityRule(fdWithout))
		assert.Equal(t, "", ExtractVisibilityRestriction(fdWithout))

		fdWith := &mockFieldDesc{opts: &descriptorpb.FieldOptions{}}
		proto.SetExtension(fdWith.opts, api_visibility.E_FieldVisibility, &api_visibility.VisibilityRule{Restriction: "INTERNAL"})
		assert.NotNil(t, GetVisibilityRule(fdWith))
		assert.Equal(t, "INTERNAL", ExtractVisibilityRestriction(fdWith))

		fdNilOpts := &mockFieldDesc{opts: nil}
		assert.Nil(t, GetVisibilityRule(fdNilOpts))
	})

	t.Run("message with visibility", func(t *testing.T) {
		mdWith := &mockMessageDesc{opts: &descriptorpb.MessageOptions{}}
		proto.SetExtension(mdWith.opts, api_visibility.E_MessageVisibility, &api_visibility.VisibilityRule{Restriction: "PREVIEW"})
		assert.Equal(t, "PREVIEW", ExtractVisibilityRestriction(mdWith))
	})

	t.Run("method with visibility", func(t *testing.T) {
		mdWith := &mockMethodDesc{opts: &descriptorpb.MethodOptions{}}
		proto.SetExtension(mdWith.opts, api_visibility.E_MethodVisibility, &api_visibility.VisibilityRule{Restriction: "INTERNAL"})
		assert.Equal(t, "INTERNAL", ExtractVisibilityRestriction(mdWith))
	})

	t.Run("service with visibility", func(t *testing.T) {
		sdWith := &mockServiceDesc{opts: &descriptorpb.ServiceOptions{}}
		proto.SetExtension(sdWith.opts, api_visibility.E_ApiVisibility, &api_visibility.VisibilityRule{Restriction: "INTERNAL,PREVIEW"})
		assert.Equal(t, "INTERNAL,PREVIEW", ExtractVisibilityRestriction(sdWith))
	})

	t.Run("enum with visibility", func(t *testing.T) {
		edWith := &mockEnumDesc{opts: &descriptorpb.EnumOptions{}}
		proto.SetExtension(edWith.opts, api_visibility.E_EnumVisibility, &api_visibility.VisibilityRule{Restriction: "ADMIN"})
		assert.Equal(t, "ADMIN", ExtractVisibilityRestriction(edWith))
	})

	t.Run("enum value with visibility", func(t *testing.T) {
		evdWith := &mockEnumValueDesc{opts: &descriptorpb.EnumValueOptions{}}
		proto.SetExtension(evdWith.opts, api_visibility.E_ValueVisibility, &api_visibility.VisibilityRule{Restriction: "ADMIN"})
		assert.Equal(t, "ADMIN", ExtractVisibilityRestriction(evdWith))
	})
}
